package controller

import (
	"context"
	"fmt"
	halkyon "halkyon.io/api/component/v1beta1"
	hLink "halkyon.io/api/link/v1beta1"
	"halkyon.io/operator/pkg/controller/framework"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

type Component struct {
	*halkyon.Component
	*framework.Requeueable
	*framework.HasDependents
}

func (in *Component) ComputeStatus(err error, helper *framework.K8SHelper) (needsUpdate bool) {
	statuses, update := in.HasDependents.ComputeStatus(in, err, helper)
	if len(in.Status.Links) > 0 {
		for i, link := range in.Status.Links {
			if link.Status == halkyon.Started {
				p, err := in.FetchUpdatedDependent(&corev1.Pod{}, helper)
				name := p.(*corev1.Pod).Name
				if err != nil || name == link.OriginalPodName {
					in.Status.Phase = halkyon.ComponentLinking
					in.SetNeedsRequeue(true)
					return false
				} else {
					// update link status
					l := &hLink.Link{}
					err := helper.Client.Get(context.TODO(), types.NamespacedName{
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
					err = helper.Client.Status().Update(context.TODO(), l)
					if err != nil {
						// todo: fix-me
						helper.ReqLogger.Error(err, "couldn't update link status", "link name", l.Name)
					}

					link.Status = halkyon.Linked
					link.OriginalPodName = ""
					in.Status.PodName = name
					in.Status.Links[i] = link // make sure we update the links with the modified value
					update = true
				}
			}
		}
	}
	// make sure we propagate the need for update even if setting the status doesn't change anything
	return in.SetSuccessStatus(statuses, "Ready") || update
}

func (in *Component) Init() bool {
	if len(in.Spec.DeploymentMode) == 0 {
		in.Spec.DeploymentMode = halkyon.DevDeploymentMode
		return true
	}
	return false
}

func (in *Component) GetAPIObject() runtime.Object {
	return in.Component
}

func NewComponent(component ...*halkyon.Component) *Component {
	c := &halkyon.Component{}
	if component != nil {
		c = component[0]
	}
	return &Component{
		Component:     c,
		Requeueable:   new(framework.Requeueable),
		HasDependents: framework.NewHasDependents(),
	}
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
	return nil // todo: implement me
}

func (in *Component) SetErrorStatus(err error) bool {
	errMsg := err.Error()
	if halkyon.ComponentFailed != in.Status.Phase || errMsg != in.Status.Message {
		in.Status.Phase = halkyon.ComponentFailed
		in.Status.Message = errMsg
		in.SetNeedsRequeue(false)
		return true
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
