package component

import (
	component "halkyon.io/api/component/v1beta1"
	"halkyon.io/api/v1beta1"
	"halkyon.io/operator-framework"
	"halkyon.io/operator/pkg"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type deployment struct {
	base
}

var _ framework.DependentResource = &deployment{}
var deploymentGVK = appsv1.SchemeGroupVersion.WithKind("Deployment")

func newDeployment(owner *component.Component) deployment {
	config := framework.NewConfig(deploymentGVK)
	config.Updated = true
	config.CheckedForReadiness = true
	d := deployment{base: newConfiguredBaseDependent(owner, config)}
	d.NameFn = d.Name
	return d
}

func (res deployment) Build(empty bool) (runtime.Object, error) {
	c := res.ownerAsComponent()
	if component.BuildDeploymentMode == c.Spec.DeploymentMode {
		return res.installBuild(empty)
	}
	return res.installDev(empty)
}

func (res deployment) Name() string {
	return pkg.DeploymentName(res.ownerAsComponent())
}

func (res deployment) GetCondition(underlying runtime.Object, err error) *v1beta1.DependentCondition {
	return framework.DefaultCustomizedGetConditionFor(res, err, underlying, func(underlying runtime.Object, cond *v1beta1.DependentCondition) {
		c := res.ownerAsComponent()
		if _, e := getImageInfo(c); e != nil {
			cond.Type = v1beta1.DependentFailed
			cond.Reason = "UnavailableRuntime"
			cond.Message = e.Error()
		}
		cond.Type = v1beta1.DependentReady
		cond.Reason = string(v1beta1.DependentReady)
		cond.Message = ""
	})
}
