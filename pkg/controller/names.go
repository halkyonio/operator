package controller

import (
	"fmt"
	halkyon "halkyon.io/api/component/v1beta1"
	"halkyon.io/operator/pkg/controller/framework"
	"halkyon.io/operator/pkg/util"
)

func PostgresName(owner framework.Resource) string {
	return framework.DefaultDependentResourceNameFor(owner)
}

func DeploymentName(c *Component) string {
	return DeploymentNameFor(c, c.Spec.DeploymentMode)
}

func DeploymentNameFor(c *Component, mode halkyon.DeploymentMode) string {
	name := framework.DefaultDependentResourceNameFor(c)
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

func ServiceAccountName(owner framework.Resource) string {
	switch owner.(type) {
	case *Capability:
		return PostgresName(owner) // todo: fix me
	case *Component:
		return "build-bot"
	default:
		panic(fmt.Sprintf("a service account shouldn't be created for '%s' %s owner", owner.GetName(), util.GetObjectName(owner)))
	}
}

func TaskName(owner framework.Resource) string {
	return "s2i-buildah-push"
}
