package capability

import (
	"encoding/gob"
	halkyon "halkyon.io/api/capability/v1beta1"
	"halkyon.io/api/v1beta1"
	"halkyon.io/operator-framework"
	capability2 "halkyon.io/operator-framework/plugins/capability"
)

// blank assignment to check that Capabilit implements Resource
var _ framework.Resource = &Capability{}

type Capability struct {
	*halkyon.Capability
	*framework.BaseResource
}

var _ framework.Resource = &Capability{}

func (in *Capability) Delete() error {
	return nil
}

func (in *Capability) CreateOrUpdate() error {
	return in.CreateOrUpdateDependents()
}

func (in *Capability) NewEmpty() framework.Resource {
	return NewCapability()
}

func (in *Capability) InitDependentResources() ([]framework.DependentResource, error) {
	c := in.Capability
	// get plugin associated with category and type
	p, err := capability2.GetPluginFor(c.Spec.Category, c.Spec.Type)
	if err != nil {
		return nil, err
	}
	return in.BaseResource.AddDependentResource(p.ReadyFor(c)...), nil
}

func (in *Capability) ComputeStatus() (needsUpdate bool) {
	statuses, notReadyWantsUpdate := in.BaseResource.ComputeStatus(in)
	return notReadyWantsUpdate || in.SetSuccessStatus(statuses, "Ready")
}

func (in *Capability) Init() bool {
	return false
}

func (in *Capability) GetAsHalkyonResource() v1beta1.HalkyonResource {
	return in.Capability
}

func NewCapability() *Capability {
	dependents := framework.NewHasDependents(&halkyon.Capability{})
	c := &Capability{
		Capability:   &halkyon.Capability{},
		BaseResource: dependents,
	}
	return c
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
	if _, err := capability2.GetPluginFor(in.Spec.Category, in.Spec.Type); err != nil {
		return err
	}
	return nil
}

func (in *Capability) SetErrorStatus(err error) bool {
	if err != nil {
		errMsg := err.Error()
		if halkyon.CapabilityFailed != in.Status.Phase || errMsg != in.Status.Message {
			in.Status.Phase = halkyon.CapabilityFailed
			in.Status.Message = errMsg
			in.SetNeedsRequeue(false)
			return true
		}
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

func init() {
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
	gob.Register(Capability{})
}
