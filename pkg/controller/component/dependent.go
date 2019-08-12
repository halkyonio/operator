package component

import (
	"github.com/halkyonio/operator/pkg/apis/component/v1alpha2"
	"github.com/halkyonio/operator/pkg/controller"
	"k8s.io/apimachinery/pkg/runtime"
)

type base struct {
	*controller.DependentResourceHelper
}

func (res base) Update(toUpdate runtime.Object) (bool, error) {
	return false, nil
}

func newBaseDependent(primaryResourceType runtime.Object, owner v1alpha2.Resource) base {
	return base{DependentResourceHelper: controller.NewDependentResource(primaryResourceType, owner)}
}

func (res base) ownerAsComponent() *v1alpha2.Component {
	return res.Owner().(*v1alpha2.Component)
}

func (res base) asComponent(object runtime.Object) *v1alpha2.Component {
	return object.(*v1alpha2.Component)
}
