package component

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"github.com/snowdrop/component-operator/pkg/controller"
	"github.com/snowdrop/component-operator/pkg/util"
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

func buildNamer(component *v1alpha2.Component) string {
	return util.DefaultDependentResourceNameFor(component) + "-build"
}
func buildOrDevNamer(c *v1alpha2.Component) string {
	return DeploymentNameFor(c, c.Spec.DeploymentMode)
}
func DeploymentNameFor(c *v1alpha2.Component, mode v1alpha2.DeploymentMode) string {
	if v1alpha2.BuildDeploymentMode == mode {
		return buildNamer(c)
	}
	return util.DefaultDependentResourceNameFor(c)
}
