package component

import (
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

func (res base) ownerAsComponent() *Component {
	return res.Owner().(*Component)
}

func (res base) asComponent(object runtime.Object) *Component {
	return object.(*Component)
}
