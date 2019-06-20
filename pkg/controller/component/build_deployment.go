package component

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"k8s.io/api/apps/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

//createBuildDeployment returns the Deployment config object to be used for deployment using a container image build by Tekton
func (r *ReconcileComponent) createBuildDeployment(res dependentResource, c *v1alpha2.Component) (runtime.Object, error) {
	ls := getAppLabels(c.Name)

	// create runtime container using built image (= created by the Tekton build task)
	runtimeContainer, err := r.getRuntimeContainerFor(c)
	if err != nil {
		return nil, err
	}
	// We will check if a Dev Deployment exists.
	// If this is the case, then that means that we are switching from dev to build mode
	// and we will enrich the deployment resource of the runtime container
	// create a "dev" version of the component to be able to check if the dev deployment exists
	devComp := c.DeepCopyObject().(*v1alpha2.Component)
	object, err := res.fetch(res, devComp)
	if err == nil {
		devDeployment := object.(*appsv1.Deployment)
		devContainer := &devDeployment.Spec.Template.Spec.Containers[0]
		runtimeContainer.Env = devContainer.Env
		runtimeContainer.EnvFrom = devContainer.EnvFrom
		runtimeContainer.Env = updateEnv(runtimeContainer.Env, c.Annotations["app.openshift.io/java-app-jar"])
		runtimeContainer.Ports = devContainer.Ports
	} else {
		// Check if Service port exists, otherwise define it
		if c.Spec.Port == 0 {
			c.Spec.Port = 8080 // Add a default port if empty
		}
		runtimeContainer.Ports = []corev1.ContainerPort{{
			ContainerPort: c.Spec.Port,
			Name:          "http",
			Protocol:      "TCP",
		}}
	}

	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "extensions/v1beta1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.name(c),
			Namespace: c.Namespace,
			Labels:    ls,
		},
		Spec: v1.DeploymentSpec{
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
		},
	}

	// Set Component instance as the owner and controller
	return dep, controllerutil.SetControllerReference(c, dep, r.Scheme)
}

func (r *ReconcileComponent) getRuntimeContainerFor(component *v1alpha2.Component) (corev1.Container, error) {
	container := corev1.Container{
		Env:             r.populatePodEnvVar(component.Spec),
		Image:           r.dockerImageURL(component),
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
