package link

import (
	"halkyon.io/api/component/v1beta1"
	v1beta12 "halkyon.io/api/link/v1beta1"
	"halkyon.io/operator-framework"
	"k8s.io/apimachinery/pkg/runtime"
)

type component struct {
	*framework.BaseDependentResource
}

func (res component) NameFrom(underlying runtime.Object) string {
	return framework.DefaultNameFrom(res, underlying)
}

func (res component) Fetch() (runtime.Object, error) {
	return framework.DefaultFetcher(res)
}

var _ framework.DependentResource = &component{}

func (res component) Update(toUpdate runtime.Object) (bool, error) {
	return false, nil
}

func newComponent(owner *v1beta12.Link) component {
	config := framework.NewConfig(v1beta1.SchemeGroupVersion.WithKind(v1beta1.Kind), owner.GetNamespace())
	config.CheckedForReadiness = true
	config.CreatedOrUpdated = false
	return component{framework.NewConfiguredBaseDependentResource(owner, config)}
}

func (res component) IsReady(underlying runtime.Object) (ready bool, message string) {
	c := underlying.(*v1beta1.Component)
	ready = c.IsReady()
	if !ready {
		message = c.Status.Message
	}
	return
}

func (res component) Name() string {
	return res.Owner().(*v1beta12.Link).Spec.ComponentName
}

func (res component) Build(empty bool) (runtime.Object, error) {
	if empty {
		return &v1beta1.Component{}, nil
	}
	// we don't want to be building anything: components are dealt with by the component controller
	return nil, nil
}
