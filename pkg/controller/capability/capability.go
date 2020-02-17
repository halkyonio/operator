package capability

import (
	"encoding/gob"
	halkyon "halkyon.io/api/capability/v1beta1"
	"halkyon.io/api/v1beta1"
	"halkyon.io/operator-framework"
	capability2 "halkyon.io/operator-framework/plugins/capability"
)

// blank assignment to check that Capability implements Resource
var _ framework.Resource = &Capability{}

type Capability struct {
	*halkyon.Capability
	*framework.BaseResource
}

func (in *Capability) GetStatus() v1beta1.Status {
	return in.Status.Status
}

func (in *Capability) SetStatus(status v1beta1.Status) {
	in.Status.Status = status
}

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
	return in.BaseResource.ComputeStatus(in)
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

func (in *Capability) CheckValidity() error {
	if _, err := capability2.GetPluginFor(in.Spec.Category, in.Spec.Type); err != nil {
		return err
	}
	return nil
}

func (in *Capability) DependentStatusFieldName() string {
	_ = in.Status.PodName // to make sure we update the value below if that field changes as returned value must match field name
	return "PodName"
}

func (in *Capability) ShouldDelete() bool {
	return true
}

func init() {
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
	gob.Register(Capability{})
}
