package component

import (
	v1beta12 "halkyon.io/api/capability/v1beta1"
	"halkyon.io/api/component/v1beta1"
	framework "halkyon.io/operator-framework"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

type capability struct {
	base
	capabilityConfig v1beta1.CapabilityConfig
}

var _ framework.DependentResource = &capability{}

func newCapability(owner *v1beta1.Component, capConfig v1beta1.CapabilityConfig) capability {
	config := framework.NewConfig(v1beta12.SchemeGroupVersion.WithKind(v1beta12.Kind))
	config.CheckedForReadiness = true
	config.CreatedOrUpdated = false
	return capability{base: newConfiguredBaseDependent(owner, config), capabilityConfig: capConfig}
}

func (res capability) Build(empty bool) (runtime.Object, error) {
	if empty {
		return &v1beta12.Capability{}, nil
	}
	// we don't want to be building anything: the capability is under halkyon's control, it's not up to the component to create it
	return nil, nil
}

func (res capability) NameFrom(underlying runtime.Object) string {
	return underlying.(*v1beta12.Capability).Name
}

func (res capability) IsReady(underlying runtime.Object) (bool, string) {
	c := underlying.(*v1beta12.Capability)
	return c.Status.Phase == v1beta12.CapabilityReady, c.Status.Message
}

func (res capability) Fetch() (runtime.Object, error) {
	if len(res.capabilityConfig.BoundTo) > 0 {
		return framework.Helper.Fetch(res.capabilityConfig.BoundTo, res.ownerAsComponent().Namespace, &v1beta12.Capability{})
	}
	return nil, errors.NewNotFound(v1beta12.SchemeGroupVersion.WithResource(v1beta12.Kind).GroupResource(), "")
}
