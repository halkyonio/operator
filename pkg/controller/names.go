package controller

import (
	"fmt"
	"halkyon.io/operator/pkg/apis/halkyon/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
)

func PostgresName(owner v1beta1.Resource) string {
	return DefaultDependentResourceNameFor(owner)
}

func DeploymentName(c *v1beta1.Component) string {
	return DeploymentNameFor(c, c.Spec.DeploymentMode)
}

func DeploymentNameFor(c *v1beta1.Component, mode v1beta1.DeploymentMode) string {
	name := DefaultDependentResourceNameFor(c)
	if v1beta1.BuildDeploymentMode == mode {
		return name + "-build"
	}
	return name
}

func PVCName(c *v1beta1.Component) string {
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
	case *v1beta1.Capability:
		return PostgresName(owner) // todo: fix me
	case *v1beta1.Component:
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
