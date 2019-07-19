package controller

import (
	"fmt"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
)

func PostgresName(owner v1alpha2.Resource) string {
	return DefaultDependentResourceNameFor(owner)
}

func DeploymentName(c *v1alpha2.Component) string {
	return DeploymentNameFor(c, c.Spec.DeploymentMode)
}

func DeploymentNameFor(c *v1alpha2.Component, mode v1alpha2.DeploymentMode) string {
	name := DefaultDependentResourceNameFor(c)
	if v1alpha2.BuildDeploymentMode == mode {
		return name + "-build"
	}
	return name
}

func PVCName(c *v1alpha2.Component) string {
	specified := c.Spec.Storage.Name
	if len(specified) > 0 {
		return specified
	}
	return "m2-data-" + c.Name // todo: use better default name?
}

func DefaultDependentResourceNameFor(owner v1alpha2.Resource) string {
	return owner.GetName()
}

func ServiceAccountName(owner v1alpha2.Resource) string {
	switch owner.(type) {
	case *v1alpha2.Capability:
		return PostgresName(owner) // todo: fix me
	case *v1alpha2.Component:
		return "build-bot"
	default:
		panic(fmt.Sprintf("a service account shouldn't be created for '%s' %s owner", owner.GetName(), GetObjectName(owner)))
	}
}

func GetObjectName(object runtime.Object) string {
	t := reflect.TypeOf(object)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

func TaskName(owner v1alpha2.Resource) string {
	return "s2i-buildah-push"
}
