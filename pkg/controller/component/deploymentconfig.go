package component

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	deploymentcfgv1 "github.com/openshift/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

//buildDeploymentConfig returns the Deployment config object
func (r *ReconcileComponent) buildDeploymentConfig(c *v1alpha2.Component) *deploymentcfgv1.DeploymentConfig {
	ls := r.getAppLabels(c.Name)
	r.populateEnvVar(c)
	dep := &deploymentcfgv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps.openshift.io/v1",
			Kind:       "DeploymentConfig",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.Name,
			Namespace: c.Namespace,
			Labels:    ls,
		},
		Spec: deploymentcfgv1.DeploymentConfigSpec{
			Replicas: int32(1),
			Strategy: deploymentcfgv1.DeploymentStrategy{
				Type: deploymentcfgv1.DeploymentStrategyTypeRolling,
			},
			Selector: ls,
			Template: &corev1.PodTemplateSpec{
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
						Image:           c.Spec.RuntimeName + ":latest",
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
							 Value: "run-java:/usr/local/s2i/run;run-node:/usr/libexec/s2i/run;compile-java:/usr/local/s2i/assemble;build:/deployments/buildapp",dd}},
						Image:                    SUPERVISOR_IMAGE_NAME + ":latest",
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
						{Name: c.Spec.Storage.Name,
							VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: c.Spec.Storage.Name}}},
					},
				}},
			Triggers: []deploymentcfgv1.DeploymentTriggerPolicy{
				{Type: deploymentcfgv1.DeploymentTriggerOnImageChange,
					ImageChangeParams: &deploymentcfgv1.DeploymentTriggerImageChangeParams{
						Automatic: true,
						ContainerNames: []string{
							c.Name,
						},
						From: corev1.ObjectReference{Kind: "ImageStreamTag", Name: c.Spec.RuntimeName + ":latest"}}},
				{Type: deploymentcfgv1.DeploymentTriggerOnImageChange,
					ImageChangeParams: &deploymentcfgv1.DeploymentTriggerImageChangeParams{
						Automatic: true,
						ContainerNames: []string{
							"copy-supervisord",
						},
						From: corev1.ObjectReference{Kind: "ImageStreamTag", Name: SUPERVISOR_IMAGE_NAME + ":latest"}}},
			},
		},
	}
	// Set Component instance as the owner and controller
	controllerutil.SetControllerReference(c, dep, r.scheme)
	return dep
}
