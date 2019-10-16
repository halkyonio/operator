package component

import (
	component "halkyon.io/api/component/v1beta1"
	"halkyon.io/operator/pkg/controller"
	"halkyon.io/operator/pkg/controller/framework"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type deployment struct {
	base
	reconciler *ComponentManager // todo: remove
}

func (res deployment) NewInstanceWith(owner framework.Resource) framework.DependentResource {
	return newOwnedDeployment(res.reconciler, owner)
}

func newDeployment(reconciler *ComponentManager) deployment {
	return newOwnedDeployment(reconciler, nil)
}

func newOwnedDeployment(reconciler *ComponentManager, owner framework.Resource) deployment {
	dependent := newBaseDependent(&appsv1.Deployment{}, owner)
	d := deployment{
		base:       dependent,
		reconciler: reconciler,
	}
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
	return controller.DeploymentName(res.ownerAsComponent())
}
