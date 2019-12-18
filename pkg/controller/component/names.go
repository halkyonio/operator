package component

import (
	halkyon "halkyon.io/api/component/v1beta1"
	"halkyon.io/api/v1beta1"
	"halkyon.io/operator-framework"
)

func DeploymentName(c *halkyon.Component) string {
	return DeploymentNameFor(c, c.Spec.DeploymentMode)
}

func DeploymentNameFor(c *halkyon.Component, mode halkyon.DeploymentMode) string {
	name := framework.DefaultDependentResourceNameFor(c)
	if halkyon.BuildDeploymentMode == mode {
		return name + "-build"
	}
	return name
}

func PVCName(c *halkyon.Component) string {
	specified := c.Spec.Storage.Name
	if len(specified) > 0 {
		return specified
	}
	return "m2-data-" + c.Name // todo: use better default name?
}

func ServiceAccountName(owner v1beta1.HalkyonResource) string {
	return "build-bot"
}

func TaskName(owner v1beta1.HalkyonResource) string {
	return "s2i-buildah-push"
}
