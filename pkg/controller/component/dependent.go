package component

import (
	component "halkyon.io/api/component/v1beta1"
	"halkyon.io/api/v1beta1"
	"halkyon.io/operator/pkg/controller"
	"k8s.io/apimachinery/pkg/runtime"
)

type base struct {
	*controller.DependentResourceHelper
}

func (res base) Update(toUpdate runtime.Object) (bool, error) {
	return false, nil
}

func newBaseDependent(primaryResourceType runtime.Object, owner v1beta1.Resource) base {
	return base{DependentResourceHelper: controller.NewDependentResource(primaryResourceType, owner)}
}

func (res base) ownerAsComponent() *component.Component {
	return res.Owner().(*component.Component)
}

func (res base) asComponent(object runtime.Object) *component.Component {
	return object.(*component.Component)
}
