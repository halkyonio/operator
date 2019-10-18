package component

import (
	component "halkyon.io/api/component/v1beta1"
	"halkyon.io/operator/pkg/controller/framework"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type deployment struct {
	base
}

func newDeployment(owner framework.Resource) deployment {
	dependent := newBaseDependent(&appsv1.Deployment{}, owner)
	d := deployment{base: dependent}
	dependent.SetDelegate(d)
	return d
}

func (res deployment) Build() (runtime.Object, error) {
	c := res.ownerAsComponent()
	if component.BuildDeploymentMode == c.Spec.DeploymentMode {
		return res.installBuild()
	}
	return res.installDev()
}

func (res deployment) Name() string {
	return DeploymentName(res.ownerAsComponent())
}
