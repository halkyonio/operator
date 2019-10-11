package component

import (
	"halkyon.io/operator/pkg/controller"
	"halkyon.io/operator/pkg/controller/framework"
	"k8s.io/apimachinery/pkg/runtime"
)

type base struct {
	*framework.DependentResourceHelper
}

func (res base) Update(toUpdate runtime.Object) (bool, error) {
	return false, nil
}

func newBaseDependent(primaryResourceType runtime.Object, owner framework.Resource) base {
	return base{DependentResourceHelper: framework.NewDependentResource(primaryResourceType, owner)}
}

func (res base) ownerAsComponent() *controller.Component {
	return res.Owner().(*controller.Component)
}

func (res base) asComponent(object runtime.Object) *controller.Component {
	return object.(*controller.Component)
}
