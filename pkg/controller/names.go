package controller

import (
	"fmt"
	capability "halkyon.io/api/capability/v1beta1"
	component "halkyon.io/api/component/v1beta1"
	"halkyon.io/api/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
)

func PostgresName(owner v1beta1.Resource) string {
	return DefaultDependentResourceNameFor(owner)
}

func DeploymentName(c *component.Component) string {
	return DeploymentNameFor(c, c.Spec.DeploymentMode)
}

func DeploymentNameFor(c *component.Component, mode component.DeploymentMode) string {
	name := DefaultDependentResourceNameFor(c)
	if component.BuildDeploymentMode == mode {
		return name + "-build"
	}
	return name
}

func PVCName(c *component.Component) string {
	specified := c.Spec.Storage.Name
	if len(specified) > 0 {
		return specified
	}
	return "m2-data-" + c.Name // todo: use better default name?
}

func DefaultDependentResourceNameFor(owner v1beta1.Resource) string {
	return owner.GetName()
}

func ServiceAccountName(owner v1beta1.Resource) string {
	switch owner.(type) {
	case *capability.Capability:
		return PostgresName(owner) // todo: fix me
	case *component.Component:
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

func TaskName(owner v1beta1.Resource) string {
	return "s2i-buildah-push"
}
