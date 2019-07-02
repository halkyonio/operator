package component

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"github.com/snowdrop/component-operator/pkg/controller"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type base struct {
	controller.BaseDependentResource
}

func newBaseDependent(primaryResourceType runtime.Object) base {
	return base{BaseDependentResource: controller.NewDependentResource(primaryResourceType)}
}

func (res base) asObject(object runtime.Object) v1.Object {
	return res.asComponent(object)
}

func (res base) ownerAsComponent() *v1alpha2.Component {
	return res.Owner().(*v1alpha2.Component)
}

func (res base) asComponent(object runtime.Object) *v1alpha2.Component {
	return object.(*v1alpha2.Component)
}

type namer func(*v1alpha2.Component) string

var defaultNamer namer = func(component *v1alpha2.Component) string {
	return component.Name
}
var buildNamer namer = func(component *v1alpha2.Component) string {
	return defaultNamer(component) + "-build"
}
var buildOrDevNamer = func(c *v1alpha2.Component) string {
	if v1alpha2.BuildDeploymentMode == c.Spec.DeploymentMode {
		return buildNamer(c)
	}
	return defaultNamer(c)
}
