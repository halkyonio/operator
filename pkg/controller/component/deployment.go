package component

import (
	component "halkyon.io/api/component/v1beta1"
	"halkyon.io/api/v1beta1"
	"halkyon.io/operator-framework"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
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
	return res.ownerAsComponent().DeploymentName()
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

func (res deployment) Update(toUpdate runtime.Object) (bool, runtime.Object, error) {
	deployment := toUpdate.(*appsv1.Deployment)
	container := deployment.Spec.Template.Spec.Containers[0]
	c := res.ownerAsComponent()
	envAsMap, err := getEnvAsMap(c)
	if err != nil {
		return false, nil, err
	}
	existingAsMap := make(map[string]string, len(container.Env))
	for _, envVar := range container.Env {
		existingAsMap[envVar.Name] = envVar.Value
	}
	if !reflect.DeepEqual(envAsMap, existingAsMap) {
		newEnvVars := make([]corev1.EnvVar, 0, len(envAsMap))
		for k, v := range envAsMap {
			newEnvVars = append(newEnvVars, corev1.EnvVar{Name: k, Value: v})
		}
		container.Env = newEnvVars
		deployment.Spec.Template.Spec.Containers[0] = container
		return true, deployment, nil
	}
	return false, deployment, nil
}

func populatePodEnvVar(component *component.Component) []corev1.EnvVar {
	tmpEnvVar, err := getEnvAsMap(component)
	if err != nil {
		panic(err)
	}

	// Convert Map to Slice
	newEnvVars := make([]corev1.EnvVar, 0, len(tmpEnvVar))
	for k, v := range tmpEnvVar {
		newEnvVars = append(newEnvVars, corev1.EnvVar{Name: k, Value: v})
	}

	return newEnvVars
}
