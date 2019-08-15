package component

import (
	component "halkyon.io/api/component/v1beta1"
	"halkyon.io/api/v1beta1"
	"halkyon.io/operator/pkg/controller"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type deployment struct {
	base
	reconciler *ReconcileComponent // todo: remove
}

func (res deployment) NewInstanceWith(owner v1beta1.Resource) controller.DependentResource {
	return newOwnedDeployment(res.reconciler, owner)
}

func newDeployment(reconciler *ReconcileComponent) deployment {
	return newOwnedDeployment(reconciler, nil)
}

func newOwnedDeployment(reconciler *ReconcileComponent, owner v1beta1.Resource) deployment {
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
