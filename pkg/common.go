package pkg

import (
	"halkyon.io/api/component/v1beta1"
	"halkyon.io/operator-framework"
)

func DeploymentName(c *v1beta1.Component) string {
	return DeploymentNameFor(c, c.Spec.DeploymentMode)
}

func DeploymentNameFor(c *v1beta1.Component, mode v1beta1.DeploymentMode) string {
	name := framework.DefaultDependentResourceNameFor(c)
	if v1beta1.BuildDeploymentMode == mode {
		return name + "-build"
	}
	return name
}
