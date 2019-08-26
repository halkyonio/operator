package controller

import (
	"fmt"
	halkyon "halkyon.io/api/component/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
)

func PostgresName(owner Resource) string {
	return DefaultDependentResourceNameFor(owner)
}

func DeploymentName(c *Component) string {
	return DeploymentNameFor(c, c.Spec.DeploymentMode)
}

func DeploymentNameFor(c *Component, mode halkyon.DeploymentMode) string {
	name := DefaultDependentResourceNameFor(c)
	if halkyon.BuildDeploymentMode == mode {
		return name + "-build"
	}
	return name
}

func PVCName(c *Component) string {
	specified := c.Spec.Storage.Name
	if len(specified) > 0 {
		return specified
	}
	return "m2-data-" + c.Name // todo: use better default name?
}

func DefaultDependentResourceNameFor(owner Resource) string {
	return owner.GetName()
}

func ServiceAccountName(owner Resource) string {
	switch owner.(type) {
	case *Capability:
		return PostgresName(owner) // todo: fix me
	case *Component:
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

func TaskName(owner Resource) string {
	return "s2i-buildah-push"
}
