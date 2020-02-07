package component

import (
	"context"
	"fmt"
	halkyon "halkyon.io/api/component/v1beta1"
	"halkyon.io/api/v1beta1"
	"halkyon.io/operator-framework"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Component struct {
	*halkyon.Component
	*framework.BaseResource
}

func (in *Component) NewEmpty() framework.Resource {
	return NewComponent()
}

func (in *Component) GetStatus() v1beta1.Status {
	return in.Status.Status
}

func (in *Component) SetStatus(status v1beta1.Status) {
	in.Status.Status = status
}

func (in *Component) InitDependentResources() ([]framework.DependentResource, error) {
	c := in.Component
	dependents := make([]framework.DependentResource, 0, 20)
	dependents = append(dependents, in.BaseResource.AddDependentResource(newRole(in), framework.NewOwnedRoleBinding(in), newServiceAccount(c), newPvc(c),
		newDeployment(c), newService(c), newRoute(c), newIngress(c), newTask(c), newTaskRun(c, in.DependentStatusFieldName()),
		newPod(c, in.DependentStatusFieldName()))...)
	requiredCapabilities := c.Spec.Capabilities.Requires
	for _, config := range requiredCapabilities {
		dependents = append(dependents, in.BaseResource.AddDependentResource(newCapability(c, config))...)
	}

	return dependents, nil
}

// blank assignment to check that Component implements Resource
var _ framework.Resource = &Component{}

func (in *Component) Delete() error {
	if framework.IsTargetClusterRunningOpenShift() {
		// Delete the ImageStream created by OpenShift if it exists as the Component doesn't own this resource
		// when it is created during build deployment mode
		imageStream := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "image.openshift.io/v1",
				"kind":       "ImageStream",
				"metadata": map[string]interface{}{
					"name":      in.GetName(),
					"namespace": in.GetNamespace(),
				},
			},
		}

		// attempt to delete the imagestream if it exists
		if e := framework.Helper.Client.Delete(context.TODO(), imageStream); e != nil && !errors.IsNotFound(e) {
			return e
		}
	}
	return nil
}

func (in *Component) CreateOrUpdate() (err error) {
	if halkyon.BuildDeploymentMode == in.Spec.DeploymentMode {
		err = in.CreateOrUpdateDependents()
	} else {
		// Enrich Component with k8s recommend Labels
		in.ObjectMeta.Labels = PopulateK8sLabels(in.Component, "Backend")

		// Enrich Env Vars with Default values
		err = populateEnvVar(in.Component)
		if err != nil {
			return err
		}

		err = in.CreateOrUpdateDependents()
	}

	// link to the capabilities if they're ready
	for _, required := range in.Spec.Capabilities.Requires {
		// get dependent corresponding to requirement
		if dependentCap, err := in.GetDependent(predicateFor(required)); err == nil {
			// retrieve the associated condition in the status and check if we're not already linked or linking the capability
			condition := in.Status.GetConditionFor(required.Name, capabilityGVK)
			if condition == nil || (condition.Type != v1beta1.DependentLinked && condition.Type != v1beta1.DependentLinking) {
				// fetch associated capability if it exists
				c, err := dependentCap.Fetch()
				if err != nil {
					return err
				}
				ready, _ := dependentCap.IsReady(c)
				if ready {
					// the capability is not already linked or linking and ready so link it
					condition.Type = v1beta1.DependentLinking
					pod, err := in.FetchUpdatedDependent(framework.TypePredicateFor(podGVK))
					if err != nil {
						return err
					}
					condition.SetAttribute("OriginalPodName", pod.(*corev1.Pod).Name)
					required.BoundTo = dependentCap.NameFrom(c)
					err = in.updateComponentWithLinkInfo(required)
					if err != nil {
						return err
					}
				}
			}
		}

	}

	if err != nil {
		return err
	}
	return nil
}

func (in *Component) updateComponentWithLinkInfo(c halkyon.CapabilityConfig) error {
	var isModified = false
	d, err := in.FetchUpdatedDependent(framework.TypePredicateFor(deploymentGVK))
	if err != nil {
		return fmt.Errorf("couldn't retrieve deployment for component '%s'", in.Name)
	}
	deployment := d.(*appsv1.Deployment)
	containers := deployment.Spec.Template.Spec.Containers
	secretName := fmt.Sprintf("%s-config", c.BoundTo) // todo: we need to retrieve the secret name from the capability

	// Check if EnvFrom already exists
	// If this is the case, exit without error
	for i := 0; i < len(containers); i++ {
		var isEnvFromExist = false
		for _, env := range containers[i].EnvFrom {
			if env.SecretRef.Name == secretName {
				// EnvFrom already exists for the Secret Ref
				isEnvFromExist = true
			}
		}
		if !isEnvFromExist {
			// Add the Secret as EnvVar to the container
			containers[i].EnvFrom = append(containers[i].EnvFrom, addSecretAsEnvFromSource(secretName))
			isModified = true
		}
	}

	if isModified {
		deployment.Spec.Template.Spec.Containers = containers
		if err := framework.Helper.Client.Update(context.TODO(), deployment); err != nil {
			// As it could be possible that we can't update the Deployment as it has been modified by another
			// process, then we will requeue
			in.SetNeedsRequeue(true)
			return err
		}
	}

	return nil
}

func addSecretAsEnvFromSource(secretName string) corev1.EnvFromSource {
	return corev1.EnvFromSource{
		SecretRef: &corev1.SecretEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{Name: secretName},
		},
	}
}

type ConfigPredicate struct {
	config halkyon.CapabilityConfig
}

func (c ConfigPredicate) Matches(resource framework.DependentResource) bool {
	capability, ok := resource.(capability)
	if !ok {
		return false
	}

	return c.config.Name == capability.capabilityConfig.Name
}

func (c ConfigPredicate) String() string {
	return selectorFor(c.config.Spec).String()
}

func predicateFor(config halkyon.CapabilityConfig) framework.Predicate {
	return ConfigPredicate{config: config}
}

func (in *Component) ComputeStatus() (needsUpdate bool) {
	needsUpdate = in.BaseResource.ComputeStatus(in)

	if len(in.Status.Conditions) > 0 {
		for i, link := range in.Status.Conditions {
			if link.Type == v1beta1.DependentLinking {
				p, err := in.FetchUpdatedDependent(framework.TypePredicateFor(podGVK))
				if err != nil || p.(*corev1.Pod).Name == link.GetAttribute("OriginalPodName") {
					in.Status.Reason = halkyon.ComponentLinking
					in.SetNeedsRequeue(true)
					return true
				} else {
					// update link status
					link.Type = v1beta1.DependentLinked
					link.SetAttribute("OriginalPodName", "")
					in.Status.PodName = p.(*corev1.Pod).Name
					in.Status.Conditions[i] = link // make sure we update the links with the modified value
					needsUpdate = true
				}
			}
		}
	}
	return needsUpdate
}

func (in *Component) Init() bool {
	if len(in.Spec.DeploymentMode) == 0 {
		in.Spec.DeploymentMode = halkyon.DevDeploymentMode
		return true
	}
	in.Spec.Storage.Name = PVCName(in.Component)
	return false
}

func (in *Component) GetAsHalkyonResource() v1beta1.HalkyonResource {
	return in.Component
}

func NewComponent() *Component {
	dependents := framework.NewHasDependents(&halkyon.Component{})
	c := &Component{
		Component:    &halkyon.Component{},
		BaseResource: dependents,
	}
	return c
}

func (in *Component) isPending() bool {
	status := halkyon.ComponentPending
	if halkyon.BuildDeploymentMode == in.Spec.DeploymentMode {
		status = halkyon.ComponentBuilding
	}
	return status == in.Status.Phase
}

func (in *Component) SetInitialStatus(msg string) bool {
	if !in.isPending() || in.Status.Message != msg {
		in.Status.Phase = halkyon.ComponentPending
		if halkyon.BuildDeploymentMode == in.Spec.DeploymentMode {
			in.Status.Phase = halkyon.ComponentBuilding
		}
		in.Status.Message = msg
		in.SetNeedsRequeue(true)
		return true
	}
	return false
}

func (in *Component) CheckValidity() error {
	// Check if Service port exists, otherwise error out
	if in.Spec.Port == 0 {
		return fmt.Errorf("component '%s' must provide a port", in.Name)
	}
	return nil
}

func (in *Component) SetErrorStatus(err error) bool {
	if err != nil {
		errMsg := err.Error()
		if halkyon.ComponentFailed != in.Status.Phase || errMsg != in.Status.Message {
			in.Status.Phase = halkyon.ComponentFailed
			in.Status.Message = errMsg
			in.SetNeedsRequeue(false)
			return true
		}
	}
	return false
}

func (in *Component) DependentStatusFieldName() string {
	_ = in.Status.PodName // to make sure we update the value below if that field changes as returned value must match field name
	return "PodName"
}

func (in *Component) SetSuccessStatus(statuses []framework.DependentResourceStatus, msg string) bool {
	// todo: compute message based on linking statuses
	changed, updatedMsg := framework.HasChangedFromStatusUpdate(&in.Status, statuses, msg)
	if changed || halkyon.ComponentReady != in.Status.Phase {
		in.Status.Phase = halkyon.ComponentReady
		in.Status.Message = updatedMsg
		in.SetNeedsRequeue(false)
		return true
	}
	return false
}

func (in *Component) GetStatusAsString() string {
	return in.Status.Phase.String()
}

func (in *Component) ShouldDelete() bool {
	return true
}

func (in *Component) Owner() v1beta1.HalkyonResource {
	return in.Component
}

func (in *Component) GetRoleName() string {
	return "image-scc-privileged-role"
}

func (in *Component) GetRoleBindingName() string {
	return "use-image-scc-privileged"
}

func (in *Component) GetAssociatedRoleName() string {
	return in.GetRoleName()
}

func (in *Component) GetServiceAccountName() string {
	return ServiceAccountName(in.Component)
}
