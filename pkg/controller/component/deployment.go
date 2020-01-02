package component

import (
	component "halkyon.io/api/component/v1beta1"
	"halkyon.io/api/v1beta1"
	"halkyon.io/operator-framework"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type deployment struct {
	base
}

var _ framework.DependentResource = &deployment{}

func newDeployment(owner v1beta1.HalkyonResource) deployment {
	return deployment{base: newBaseDependent(&appsv1.Deployment{}, owner)}
}

func (res deployment) Build(empty bool) (runtime.Object, error) {
	c := res.ownerAsComponent()
	if component.BuildDeploymentMode == c.Spec.DeploymentMode {
		return res.installBuild(empty)
	}
	return res.installDev(empty)
}

func (res deployment) Name() string {
	return DeploymentName(res.ownerAsComponent())
}
