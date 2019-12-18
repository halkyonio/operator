package component

import (
	"halkyon.io/api/component/v1beta1"
	"halkyon.io/operator-framework"
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

func (res base) ownerAsComponent() *v1beta1.Component {
	return res.Owner().(*v1beta1.Component)
}

func (res base) asComponent(object runtime.Object) *Component {
	return object.(*Component)
}
