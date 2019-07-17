package component

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"github.com/snowdrop/component-operator/pkg/controller"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type deployment struct {
	base
	reconciler *ReconcileComponent // todo: remove
}

func (res deployment) NewInstanceWith(owner v1alpha2.Resource) controller.DependentResource {
	return newOwnedDeployment(res.reconciler, owner)
}

func newDeployment(reconciler *ReconcileComponent) deployment {
	return newOwnedDeployment(reconciler, nil)
}

func newOwnedDeployment(reconciler *ReconcileComponent, owner v1alpha2.Resource) deployment {
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
	if v1alpha2.BuildDeploymentMode == c.Spec.DeploymentMode {
		return res.installBuild()
	}
	return res.installDev()
}

func (res deployment) Name() string {
	return buildOrDevNamer(res.ownerAsComponent())
}
