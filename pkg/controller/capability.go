package controller

import (
	"fmt"
	halkyon "halkyon.io/api/capability/v1beta1"
	"halkyon.io/operator/pkg/controller/framework"
	"k8s.io/apimachinery/pkg/runtime"
)

type Capability struct {
	*halkyon.Capability
	*framework.Requeueable
	*framework.HasDependents
}

func (in *Capability) ComputeStatus(err error, helper *framework.K8SHelper) (needsUpdate bool) {
	statuses, update := in.HasDependents.ComputeStatus(in, err, helper)
	return in.SetSuccessStatus(statuses, "Ready") || update
}

func (in *Capability) Init() bool {
	return false
}

func (in *Capability) GetAPIObject() runtime.Object {
	return in.Capability
}

func NewCapability(capability ...*halkyon.Capability) *Capability {
	c := &halkyon.Capability{}
	if capability != nil {
		c = capability[0]
	}
	return &Capability{
		Capability:    c,
		Requeueable:   new(framework.Requeueable),
		HasDependents: new(framework.HasDependents),
	}
}

func (in *Capability) SetInitialStatus(msg string) bool {
	if halkyon.CapabilityPending != in.Status.Phase || in.Status.Message != msg {
		in.Status.Phase = halkyon.CapabilityPending
		in.Status.Message = msg
		in.SetNeedsRequeue(true)
		return true
	}
	return false
}

func (in *Capability) CheckValidity() error {
	if !halkyon.DatabaseCategory.Equals(in.Spec.Category) {
		return fmt.Errorf("unsupported '%s' capability category", in.Spec.Category)
	}
	if !halkyon.PostgresType.Equals(in.Spec.Type) {
		return fmt.Errorf("unsupported '%s' database type", in.Spec.Type)
	}
	return nil
}

func (in *Capability) SetErrorStatus(err error) bool {
	errMsg := err.Error()
	if halkyon.CapabilityFailed != in.Status.Phase || errMsg != in.Status.Message {
		in.Status.Phase = halkyon.CapabilityFailed
		in.Status.Message = errMsg
		in.SetNeedsRequeue(false)
		return true
	}
	return false
}

func (in *Capability) DependentStatusFieldName() string {
	_ = in.Status.PodName // to make sure we update the value below if that field changes as returned value must match field name
	return "PodName"
}

func (in *Capability) SetSuccessStatus(statuses []framework.DependentResourceStatus, msg string) bool {
	changed, updatedMsg := framework.HasChangedFromStatusUpdate(&in.Status, statuses, msg)
	if changed || halkyon.CapabilityReady != in.Status.Phase {
		in.Status.Phase = halkyon.CapabilityReady
		in.Status.Message = updatedMsg
		in.SetNeedsRequeue(false)
		return true
	}
	return false
}

func (in *Capability) GetStatusAsString() string {
	return in.Status.Phase.String()
}

func (in *Capability) ShouldDelete() bool {
	return true
}

func (in *Capability) Delete() error {
	return nil
}

func (in *Capability) SetPrimaryResourceStatus(statuses []framework.DependentResourceStatus) bool {
	return in.SetSuccessStatus(statuses, "Ready")
}
