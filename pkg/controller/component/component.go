package component

import (
	"context"
	goerrors "errors"
	"fmt"
	halkyon "halkyon.io/api/component/v1beta1"
	"halkyon.io/api/v1beta1"
	"halkyon.io/operator-framework"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// blank assignment to check that Component implements Resource
var _ framework.Resource = &Component{}

// Component implements the Resource interface to handle behavior tied to the state of Halkyon's Component CR.
type Component struct {
	*halkyon.Component
	*framework.BaseResource
}

// NewComponent creates a new Component instance, reusing BaseResource as the foundation for its behavior
func NewComponent() *Component {
	c := &Component{Component: &halkyon.Component{}}
	// initialize the BaseResource, delegating its status handling to our newly created instance as StatusAware instance
	c.BaseResource = framework.NewBaseResource(c)
	c.Component.SetGroupVersionKind(c.Component.GetGroupVersionKind()) // make sure that GVK is set on the runtime object
	return c
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
		newDeployment(c), newService(c), newRoute(c), newIngress(c), newTask(c), newTaskRun(c), newPod(c))...)

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

func (in *Component) ProvideDefaultValues() bool {
	if len(in.Spec.DeploymentMode) == 0 {
		in.Spec.DeploymentMode = halkyon.DevDeploymentMode
		return true
	}
	in.Spec.Storage.Name = PVCName(in.Component)
	return false
}

func (in *Component) GetUnderlyingAPIResource() framework.SerializableResource {
	return in.Component
}

func (in *Component) CheckValidity() error {
	// Check if Service port exists, otherwise error out
	if in.Spec.Port == 0 {
		return fmt.Errorf("component '%s' must provide a port", in.Name)
	}
	return nil
}

func (in *Component) Owner() framework.SerializableResource {
	return in.GetUnderlyingAPIResource()
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

func (in *Component) Handle(err error) (bool, v1beta1.Status) {
	// unwrap error to check if we have contract error
	unwrapped := goerrors.Unwrap(err)
	if unwrapped != nil {
		if _, ok := unwrapped.(*contractError); ok {
			msg := unwrapped.Error()
			// if we have a contract error but the pod is ready, set the status to PushReady
			if dependent, e := in.GetDependent(framework.TypePredicateFor(halkyon.PodGVK)); e == nil {
				condition := dependent.GetCondition(dependent.Fetch())
				updated := in.Status.SetCondition(condition) // set the condition on the status to make sure we record the pod name
				if updated {
					if condition.IsReady() && in.Status.Reason != halkyon.PushReady {
						in.Status.Reason = halkyon.PushReady
						in.Status.Message = msg
					}
					return updated, in.Status.Status
				}
			}
			return false, in.Status.Status
		}
	}

	// if we don't have a contract error, proceed with the default handler
	updated, status := framework.DefaultErrorHandler(in.Status.Status, err)
	in.Status.Status = status
	return updated, status
}
