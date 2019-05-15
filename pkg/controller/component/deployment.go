package component

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"github.com/snowdrop/component-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

//buildDeployment returns the Deployment config object
func (r *ReconcileComponent) buildDeployment(c *v1alpha2.Component) *v1beta1.Deployment {
	ls := r.getAppLabels(c.Name)
	dep := &v1beta1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "extensions/v1beta1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.Name,
			Namespace: c.Namespace,
			Labels:    ls,
		},
		Spec: v1beta1.DeploymentSpec{
			Strategy: v1beta1.DeploymentStrategy{
				Type: v1beta1.RollingUpdateDeploymentStrategyType,
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
					Containers: []corev1.Container{{
						Args: []string{
							"-c",
							"/var/lib/supervisord/conf/supervisor.conf",
						},
						Command: []string{
							"/var/lib/supervisord/bin/supervisord",
						},
						Env:             *r.populatePodEnvVar(c),
						Image:           c.Spec.GetImageReference(),
						ImagePullPolicy: corev1.PullAlways,
						Name:            c.Name,
						Ports: []corev1.ContainerPort{{
							ContainerPort: c.Spec.Port,
							Name:          "http",
							Protocol:      "TCP",
						}},
						VolumeMounts: []corev1.VolumeMount{
							{Name: "shared-data", MountPath: "/var/lib/supervisord"},
							{Name: c.Spec.Storage.Name, MountPath: "/tmp/artifacts"},
						},
					}},
					InitContainers: []corev1.Container{{
						Env: []corev1.EnvVar{
							{Name: "CMDS", Value: ""}},
						Image:                    util.GetImageReference(SUPERVISOR_IMAGE_NAME),
						ImagePullPolicy:          corev1.PullAlways,
						Name:                     SUPERVISOR_IMAGE_NAME,
						TerminationMessagePath:   "dev/termination-log",
						TerminationMessagePolicy: "File",
						VolumeMounts: []corev1.VolumeMount{
							{Name: "shared-data", MountPath: "/var/lib/supervisord"},
						},
					}},
					Volumes: []corev1.Volume{
						{Name: "shared-data",
							VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
						{Name: "",
							VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: ""}}},
					},
				}},
		},
	}

	// Set Component instance as the owner and controller
	controllerutil.SetControllerReference(c, dep, r.scheme)
	return dep
}
