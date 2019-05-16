package component

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"github.com/snowdrop/component-operator/pkg/util"
	"k8s.io/api/apps/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

//buildDeployment returns the Deployment config object
func (r *ReconcileComponent) buildDeployment(c *v1alpha2.Component) *appsv1.Deployment {
	ls := r.getAppLabels(c.Name)
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
							{Name: "CMDS",
								Value: "run-java:/usr/local/s2i/run;run-node:/usr/libexec/s2i/run;compile-java:/usr/local/s2i/assemble;build:/deployments/buildapp"}},
						Image:                    util.GetImageReference(SupervisorImageName),
						ImagePullPolicy:          corev1.PullAlways,
						Name:                     SupervisorImageName,
						TerminationMessagePath:   "dev/termination-log",
						TerminationMessagePolicy: "File",
						VolumeMounts: []corev1.VolumeMount{
							{Name: "shared-data", MountPath: "/var/lib/supervisord"},
						},
					}},
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
	controllerutil.SetControllerReference(c, dep, r.scheme)
	return dep
}
