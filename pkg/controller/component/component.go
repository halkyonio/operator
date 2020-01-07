package component

import (
	"context"
	"fmt"
	halkyon "halkyon.io/api/component/v1beta1"
	hLink "halkyon.io/api/link/v1beta1"
	"halkyon.io/api/v1beta1"
	"halkyon.io/operator-framework"
	"halkyon.io/operator-framework/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

type Component struct {
	*halkyon.Component
	*framework.BaseResource
}

func (in *Component) InitDependents() []framework.DependentResource {
	c := in.Component
	res := []framework.DependentResource{newRole(c), newRoleBinding(c), newServiceAccount(c), newPvc(c), newDeployment(c),
		newService(c), newRoute(c), newIngress(c), newTask(c), newTaskRun(c, in.DependentStatusFieldName()), newPod(c, in.DependentStatusFieldName())}
	in.BaseResource.AddDependentResource(res...)
	return res
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
		populateEnvVar(in.Component)

		return in.CreateOrUpdateDependents()
	}
	return err
}

func (in *Component) FetchAndCreateNew(name, namespace string) (framework.Resource, error) {
	return in.BaseResource.FetchAndInitNewResource(name, namespace, NewComponent())
}

func (in *Component) ComputeStatus() (needsUpdate bool) {
	statuses, notReadyWantsUpdate := in.BaseResource.ComputeStatus(in)
	if len(in.Status.Links) > 0 {
		for i, link := range in.Status.Links {
			if link.Status == halkyon.Started {
				p, err := in.FetchUpdatedDependent(util.GetObjectName(&corev1.Pod{}))
				if err != nil || p.(*corev1.Pod).Name == link.OriginalPodName {
					in.Status.Phase = halkyon.ComponentLinking
					in.SetNeedsRequeue(true)
					return false
				} else {
					// update link status
					l := &hLink.Link{}
					err := framework.Helper.Client.Get(context.TODO(), types.NamespacedName{
						Namespace: in.Namespace,
						Name:      link.Name,
					}, l)
					if err != nil {
						// todo: is this appropriate?
						link.Status = halkyon.Errored
						in.Status.Message = fmt.Sprintf("couldn't retrieve '%s' link", link.Name)
						return true
					}

					l.Status.Message = fmt.Sprintf("'%s' finished linking", in.Name)
					err = framework.Helper.Client.Status().Update(context.TODO(), l)
					if err != nil {
						// todo: fix-me
						panic(fmt.Errorf("couldn't update %s link status: %v", l.Name, err))
					}

					link.Status = halkyon.Linked
					link.OriginalPodName = ""
					in.Status.PodName = p.(*corev1.Pod).Name
					in.Status.Links[i] = link // make sure we update the links with the modified value
					notReadyWantsUpdate = true
				}
			}
		}
	}
	return notReadyWantsUpdate || in.SetSuccessStatus(statuses, "Ready")
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
