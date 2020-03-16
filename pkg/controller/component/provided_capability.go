package component

import (
	"fmt"
	v1beta12 "halkyon.io/api/capability/v1beta1"
	"halkyon.io/api/component/v1beta1"
	beta1 "halkyon.io/api/v1beta1"
	framework "halkyon.io/operator-framework"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type providedCapability struct {
	base
	capabilityConfig v1beta1.CapabilityConfig
}

var _ framework.DependentResource = &providedCapability{}

func newProvidedCapability(owner *v1beta1.Component, capConfig v1beta1.CapabilityConfig) providedCapability {
	config := framework.NewConfig(capabilityGVK)
	config.TypeName = "Provided Capability"
	c := providedCapability{base: newConfiguredBaseDependent(owner, config), capabilityConfig: capConfig}
	c.NameFn = c.Name
	return c
}

func (res providedCapability) Build(empty bool) (runtime.Object, error) {
	capability := &v1beta12.Capability{}
	if !empty {
		c := res.ownerAsComponent()
		ls := getAppLabels(c)
		capability.ObjectMeta = v1.ObjectMeta{
			Name:      res.Name(),
			Namespace: c.Namespace,
			Labels:    ls,
		}
		capability.Spec = res.capabilityConfig.Spec

		v1beta1.AddCapabilityParameterIfNeeded(beta1.NameValuePair{Name: v1beta1.TargetComponentDefaultParameterName, Value: c.GetName()}, capability)
		v1beta1.AddCapabilityParameterIfNeeded(beta1.NameValuePair{Name: v1beta1.TargetPortDefaultParameterName, Value: fmt.Sprintf("%d", c.Spec.Port)}, capability)
	}
	return capability, nil
}

func (res providedCapability) Name() string {
	return res.capabilityConfig.Name
}

func (res providedCapability) GetCondition(_ runtime.Object, _ error) *beta1.DependentCondition {
	panic("should never be called")
}

func (res providedCapability) Fetch() (runtime.Object, error) {
	return framework.DefaultFetcher(res)
}
