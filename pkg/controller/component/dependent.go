package component

import (
	"halkyon.io/operator/pkg/controller"
	"k8s.io/apimachinery/pkg/runtime"
)

type base struct {
	*controller.DependentResourceHelper
}

func (res base) Update(toUpdate runtime.Object) (bool, error) {
	return false, nil
}

func newBaseDependent(primaryResourceType runtime.Object, owner controller.Resource) base {
	return base{DependentResourceHelper: controller.NewDependentResource(primaryResourceType, owner)}
}

func (res base) ownerAsComponent() *controller.Component {
	return res.Owner().(*controller.Component)
}

func (res base) asComponent(object runtime.Object) *controller.Component {
	return object.(*controller.Component)
}
