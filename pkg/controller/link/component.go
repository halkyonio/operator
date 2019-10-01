package link

import (
	"halkyon.io/api/component/v1beta1"
	"halkyon.io/operator/pkg/controller"
	"k8s.io/apimachinery/pkg/runtime"
)

type component struct {
	*controller.DependentResourceHelper
}

func (res component) Update(toUpdate runtime.Object) (bool, error) {
	return false, nil
}

func (res component) NewInstanceWith(owner controller.Resource) controller.DependentResource {
	return newOwnedComponent(owner)
}

func newComponent() component {
	return newOwnedComponent(nil)
}

func newOwnedComponent(owner controller.Resource) component {
	resource := controller.NewDependentResource(&v1beta1.Component{}, owner)
	c := component{DependentResourceHelper: resource}
	resource.SetDelegate(c)
	return c
}

func (component) ShouldBeCheckedForReadiness() bool {
	return true
}

func (res component) IsReady(underlying runtime.Object) bool {
	c := underlying.(*v1beta1.Component)
	return v1beta1.ComponentReady == c.Status.Phase || v1beta1.ComponentRunning == c.Status.Phase
}

func (res component) Name() string {
	return res.Owner().(*controller.Link).Spec.ComponentName
}

func (res component) Build() (runtime.Object, error) {
	// we don't want to be building anything: components are dealt with by the component controller
	return nil, nil
}

func (res component) CanBeCreatedOrUpdated() bool {
	return false
}
