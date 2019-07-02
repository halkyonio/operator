package component

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type deployment struct {
	base
	reconciler *ReconcileComponent // todo: remove
}

func newDeployment(reconciler *ReconcileComponent) deployment {
	return deployment{
		base:       newBaseDependent(&appsv1.Deployment{}),
		reconciler: reconciler,
	}
}

func (res deployment) Build() (runtime.Object, error) {
	c := res.ownerAsComponent()
	if v1alpha2.BuildDeploymentMode == c.Spec.DeploymentMode {
		if err := res.reconciler.setInitialStatus(c, v1alpha2.ComponentBuilding); err != nil {
			return nil, err
		}
		return res.installBuild()
	} else if err := res.reconciler.setInitialStatus(c, v1alpha2.ComponentPending); err != nil {
		return nil, err
	}
	return res.installDev()
}

func (res deployment) Name() string {
	c := res.ownerAsComponent()
	if v1alpha2.BuildDeploymentMode == c.Spec.DeploymentMode {
		return buildNamer(c)
	}
	return defaultNamer(c)
}
