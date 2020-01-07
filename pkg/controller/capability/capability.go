package capability

import (
	"encoding/gob"
	"fmt"
	halkyon "halkyon.io/api/capability/v1beta1"
	"halkyon.io/api/v1beta1"
	"halkyon.io/operator-framework"
	capability2 "halkyon.io/plugins/capability"
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

func (in *Capability) FetchAndCreateNew(name, namespace string, callback framework.WatchCallback) (framework.Resource, error) {
	return framework.FetchAndInitNewResource(name, namespace, newEmptyCapability(), callback, func(toInit v1beta1.HalkyonResource) ([]framework.DependentResource, error) {
		c := toInit.(*halkyon.Capability)
		// get plugin associated with category and type
		category := c.Spec.Category
		capabilityType := c.Spec.Type
		for _, p := range capability2.Plugins {
			if p.GetCategory() == category && p.GetType() == capabilityType {
				// init dependents for given capability type
				return p.ReadyFor(c), nil
			}
		}
		return nil, fmt.Errorf("couldn't find a plugin to handle capability with category %s and type %s", category, capabilityType)
	})
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
	return newEmptyCapability()
}

func newEmptyCapability() *Capability {
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
	category := in.Spec.Category
	types := capability2.SupportedCategories[category]
	if len(types) == 0 {
		return fmt.Errorf("unsupported '%s' capability category", category)
	}
	t := in.Spec.Type
	if !types[t] {
		return fmt.Errorf("unsupported '%s' type for '%s'", t, category)
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
