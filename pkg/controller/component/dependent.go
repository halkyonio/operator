package component

import (
	"github.com/halkyonio/operator/pkg/apis/halkyon/v1beta1"
	"github.com/halkyonio/operator/pkg/controller"
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

func (res base) ownerAsComponent() *v1beta1.Component {
	return res.Owner().(*v1beta1.Component)
}

func (res base) asComponent(object runtime.Object) *v1beta1.Component {
	return object.(*v1beta1.Component)
}
