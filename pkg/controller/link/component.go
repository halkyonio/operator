package link

import (
	"halkyon.io/api/component/v1beta1"
	"halkyon.io/operator/pkg/controller/framework"
	"k8s.io/apimachinery/pkg/runtime"
)

type component struct {
	*framework.DependentResourceHelper
}

func (res component) Update(toUpdate runtime.Object) (bool, error) {
	return false, nil
}

func newComponent(owner framework.Resource) component {
	resource := framework.NewDependentResource(&v1beta1.Component{}, owner)
	c := component{DependentResourceHelper: resource}
	resource.SetDelegate(c)
	return c
}

func (component) ShouldBeCheckedForReadiness() bool {
	return true
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
	return res.Owner().(*Link).Spec.ComponentName
}

func (res component) Build() (runtime.Object, error) {
	// we don't want to be building anything: components are dealt with by the component controller
	return nil, nil
}

func (res component) CanBeCreatedOrUpdated() bool {
	return false
}
