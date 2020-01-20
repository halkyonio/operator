package component

import (
	component "halkyon.io/api/component/v1beta1"
	framework "halkyon.io/operator-framework"
	"halkyon.io/operator/pkg"
	"k8s.io/api/apps/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	// "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

//createBuildDeployment returns the Deployment config object to be used for deployment using a container image build by Tekton
func (res deployment) installBuild(empty bool) (runtime.Object, error) {
	dep := &appsv1.Deployment{}
	if !empty {
		c := res.ownerAsComponent()
		ls := getAppLabels(c)

		// create runtime container using built image (= created by the Tekton build task)
		runtimeContainer, err := getRuntimeContainerFor(c)
		if err != nil {
			return nil, err
		}
		// We will check if a Dev Deployment exists.
		// If this is the case, then that means that we are switching from dev to build mode
		// and we will enrich the deployment resource of the runtime container
		// create a "dev" version of the component to be able to check if the dev deployment exists
		devDeployment := &appsv1.Deployment{}
		_, err = framework.Helper.Fetch(pkg.DeploymentNameFor(c, component.DevDeploymentMode), c.Namespace, devDeployment)
		if err == nil {
			devContainer := &devDeployment.Spec.Template.Spec.Containers[0]
			runtimeContainer.Env = devContainer.Env
			runtimeContainer.EnvFrom = devContainer.EnvFrom
			runtimeContainer.Env = updateEnv(runtimeContainer.Env, c.Annotations["app.openshift.io/java-app-jar"])
			runtimeContainer.Ports = devContainer.Ports
		} else {
			runtimeContainer.Ports = []corev1.ContainerPort{{
				ContainerPort: c.Spec.Port,
				Name:          "http",
				Protocol:      "TCP",
			}}
		}

		dep.ObjectMeta = metav1.ObjectMeta{
			Name:      res.Name(),
			Namespace: c.Namespace,
			Labels:    ls,
		}
		dep.Spec = v1.DeploymentSpec{
			Strategy: v1.DeploymentStrategy{
				Type: v1.RollingUpdateDeploymentStrategyType,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
					Name:   c.Name,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{runtimeContainer},
				}},
		}

	}

	// Set Component instance as the owner and controller
	return dep, nil
}

func getRuntimeContainerFor(component *component.Component) (corev1.Container, error) {
	container := corev1.Container{
		Env:             populatePodEnvVar(component),
		Image:           dockerImageURL(component),
		ImagePullPolicy: corev1.PullAlways,
		Name:            component.Name,
	}
	return container, nil
}

func updateEnv(envs []corev1.EnvVar, jarName string) []corev1.EnvVar {
	newEnvs := []corev1.EnvVar{}
	for _, s := range envs {
		if s.Name == "JAVA_APP_JAR" {
			newEnvs = append(newEnvs, corev1.EnvVar{Name: s.Name, Value: jarName})
		} else {
			newEnvs = append(newEnvs, s)
		}
	}
	return newEnvs
}
