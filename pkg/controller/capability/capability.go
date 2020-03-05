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

func (in *Capability) ProvideDefaultValues() bool {
	return false
}

func (in *Capability) GetUnderlyingAPIResource() framework.SerializableResource {
	return in.Capability
}

func NewCapability() *Capability {
	c := &Capability{
		Capability:   &halkyon.Capability{},
		BaseResource: framework.NewBaseResource(),
	}
	c.Capability.SetGroupVersionKind(c.Capability.GetGroupVersionKind()) // make sure that GVK is set on the runtime object
	return c
}

func (in *Capability) CheckValidity() error {
	plugin, err := capability2.GetPluginFor(in.Spec.Category, in.Spec.Type)
	if err != nil {
		return err
	}
	return plugin.CheckValidity(in.Capability)
}

func (in *Capability) Handle(err error) (bool, v1beta1.Status) {
	return framework.DefaultErrorHandler(in.Status.Status, err)
}

func init() {
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
	gob.Register(Capability{})
}
