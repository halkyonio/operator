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
func (r *ReconcileComponent) createBuildDeployment(c *v1alpha2.Component) (runtime.Object, error) {
	ls := r.getAppLabels(c.Name)

	// create runtime container
	runtimeContainer, err := r.getBaseContainerFor(c)
	if err != nil {
		return nil, err
	}
	runtimeContainer.Ports = []corev1.ContainerPort{{
		ContainerPort: c.Spec.Port,
		Name:          "http",
		Protocol:      "TCP",
	}}
	runtimeContainer.Image = r.dockerImageURL(c)
	runtimeContainer.VolumeMounts = append(runtimeContainer.VolumeMounts, corev1.VolumeMount{Name: c.Spec.Storage.Name, MountPath: "/tmp/artifacts"})

	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "extensions/v1beta1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.Name,
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
					Volumes:    []corev1.Volume{
						{Name: "shared-data",
							VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
						{Name: c.Spec.Storage.Name,
							VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: c.Spec.Storage.Name}}},
					},
				}},
		},
	}

	// Set Component instance as the owner and controller
	return dep, controllerutil.SetControllerReference(c, dep, r.scheme)
}
