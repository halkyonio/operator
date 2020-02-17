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
		dependents = append(dependents, in.BaseResource.AddDependentResource(newRequiredCapability(c, config))...)
	}

	providedCapabilities := c.Spec.Capabilities.Provides
	for _, config := range providedCapabilities {
		dependents = append(dependents, in.BaseResource.AddDependentResource(newProvidedCapability(c, config))...)
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

	if err != nil {
		return err
	}

	// link to the capabilities if they're ready and we've bound them to a capability already
	needsSpecUpdate := false
	defer func() {
		if needsSpecUpdate {
			_ = framework.Helper.Client.Update(context.Background(), in.Component)
		}
	}()
	for i, required := range in.Spec.Capabilities.Requires {
		if dependentCap, err := in.GetDependent(predicateFor(required.CapabilityConfig)); err == nil {
			// attempt to retrieve the associated capability, this will bound the capability if set to auto-bindable
			c, err := dependentCap.Fetch()

			// if required wasn't bound and it is now, we need to update the spec
			// "re-fetch" current required since it's been updated when the capability auto-bounded when fetched, hackish I know
			refreshedRequired := in.Spec.Capabilities.Requires[i]
			if required.BoundTo != refreshedRequired.BoundTo {
				required = refreshedRequired
				needsSpecUpdate = true
			}

			// if the capability is bound and ready
			condition := dependentCap.GetCondition(c, err)
			if len(required.BoundTo) > 0 && condition.IsReady() {
				// check if the capability is already linked by checking if the associated deployment has been updated
				updatedDeployment, err := in.updateComponentWithLinkInfo(required)
				if err != nil {
					return err
				}
				// if updated deployment exists, we are not linked yet
				if updatedDeployment != nil {
					// mark the component as linking
					in.Status.Reason = halkyon.ComponentLinking // todo: do we need to track pod name to check if linking is done?

					// send updated deployment
					if err := framework.Helper.Client.Update(context.Background(), updatedDeployment); err != nil {
						// As it could be possible that we can't update the Deployment as it has been modified by another
						// process, then we will requeue
						in.SetNeedsRequeue(true)
						return err
					}
				}
			}
		}
	}

	return
}

func (in *Component) updateComponentWithLinkInfo(c halkyon.RequiredCapabilityConfig) (updatedDeployment *appsv1.Deployment, err error) {
	d, err := in.FetchUpdatedDependent(framework.TypePredicateFor(deploymentGVK))
	if err != nil {
		return nil, fmt.Errorf("couldn't retrieve deployment for component '%s'", in.Name)
	}
	deployment := d.(*appsv1.Deployment)
	containers := deployment.Spec.Template.Spec.Containers
	secretName := fmt.Sprintf("%s-config", c.BoundTo) // todo: we need to retrieve the secret name from the capability

	// Check if EnvFrom already exists
	// If this is the case, exit without error
	isModified := false
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
		updatedDeployment = deployment
	}

	return
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
	capability, ok := resource.(requiredCapability)
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

	/*if len(in.Status.Conditions) > 0 {
		for i, dependentCondition := range in.Status.Conditions {
			if dependentCondition.Type == v1beta1.DependentLinking {
				p, err := in.FetchUpdatedDependent(framework.TypePredicateFor(podGVK))
				if err != nil || p.(*corev1.Pod).Name == dependentCondition.GetAttribute("OriginalPodName") {
					in.Status.Reason = halkyon.ComponentLinking
					in.SetNeedsRequeue(true)
					return true
				} else {
					// update link status
					dependentCondition.Type = v1beta1.DependentLinked
					dependentCondition.SetAttribute("OriginalPodName", "")
					in.Status.PodName = p.(*corev1.Pod).Name
					in.Status.Conditions[i] = dependentCondition // make sure we update the links with the modified value
					needsUpdate = true
				}
			} else {
				// if pod is ready, update pod name status field
				if dependentCondition.DependentType == podGVK && dependentCondition.Reason == v1beta1.ReasonReady {
					podName := dependentCondition.GetAttribute("PodName")
					if in.Status.PodName != podName {
						in.Status.PodName = podName
						needsUpdate = true
					}
				}
			}
		}
	}*/

	return
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

func (in *Component) CheckValidity() error {
	// Check if Service port exists, otherwise error out
	if in.Spec.Port == 0 {
		return fmt.Errorf("component '%s' must provide a port", in.Name)
	}
	return nil
}

func (in *Component) DependentStatusFieldName() string {
	_ = in.Status.PodName // to make sure we update the value below if that field changes as returned value must match field name
	return "PodName"
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
