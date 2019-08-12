package component

import (
	"fmt"
	"github.com/halkyonio/operator/pkg/apis/component/v1alpha2"
	"github.com/halkyonio/operator/pkg/util"
	"k8s.io/api/apps/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	supervisorContainerName = "copy-supervisord"
	supervisorImageId       = "supervisord"
	latestVersionTag        = "latest"
)

//buildDevDeployment returns the Deployment config object
func (res deployment) installDev() (runtime.Object, error) {
	c := res.ownerAsComponent()
	ls := getAppLabels(c.Name)

	// create runtime container
	r := res.reconciler
	runtimeContainer, err := r.getBaseContainerFor(c)
	if err != nil {
		return nil, err
	}
	runtimeContainer.Args = []string{
		"-c",
		"/var/lib/supervisord/conf/supervisor.conf",
	}
	runtimeContainer.Command = []string{"/var/lib/supervisord/bin/supervisord"}
	runtimeContainer.Ports = []corev1.ContainerPort{{
		ContainerPort: c.Spec.Port,
		Name:          "http",
		Protocol:      "TCP",
	}}
	runtimeContainer.VolumeMounts = append(runtimeContainer.VolumeMounts, corev1.VolumeMount{Name: c.Spec.Storage.Name, MountPath: "/tmp/artifacts"})

	// create the supervisor init container
	supervisorContainer, err := r.getBaseContainerFor(r.supervisor)
	if err != nil {
		return nil, err
	}
	supervisorContainer.TerminationMessagePath = "/dev/termination-log"
	supervisorContainer.TerminationMessagePolicy = "File"

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.Name(),
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
					Containers:     []corev1.Container{runtimeContainer},
					InitContainers: []corev1.Container{supervisorContainer},
					Volumes: []corev1.Volume{
						{Name: "shared-data",
							VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
						{Name: c.Spec.Storage.Name,
							VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: c.Spec.Storage.Name}}},
					},
				}},
		},
	}

	// Set Component instance as the owner and controller
	return dep, nil
}

func (r *ReconcileComponent) getBaseContainerFor(component *v1alpha2.Component) (corev1.Container, error) {
	runtimeImage, err := r.getImageReference(component.Spec)
	if err != nil {
		return corev1.Container{}, err
	}

	container := corev1.Container{
		Env:             r.populatePodEnvVar(component.Spec),
		Image:           runtimeImage,
		ImagePullPolicy: corev1.PullAlways,
		Name:            component.Name,
		VolumeMounts: []corev1.VolumeMount{
			{Name: "shared-data", MountPath: "/var/lib/supervisord"},
		},
	}
	return container, nil
}

func (r *ReconcileComponent) getImageInfo(component v1alpha2.ComponentSpec) (imageInfo, error) {
	image, ok := r.runtimeImages[component.Runtime]
	if !ok {
		return imageInfo{}, fmt.Errorf("unknown image identifier: %s", component.Runtime)
	}
	return image, nil
}

func (r *ReconcileComponent) getImageReference(component v1alpha2.ComponentSpec) (string, error) {
	image, err := r.getImageInfo(component)
	if err != nil {
		return "", err
	}

	// todo: compute image version based on runtime version if needed
	v := latestVersionTag

	return util.GetImageReference(image.registryRef, v), nil
}

func (r *ReconcileComponent) populatePodEnvVar(component v1alpha2.ComponentSpec) []corev1.EnvVar {
	tmpEnvVar, err := r.getEnvAsMap(component)
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
