package component

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	v1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	taskS2iBuildahPusName = "s2i-buildah-push"
	serviceAccountName    = "build-bot"
)

var (
	gitInputs = &v1alpha1.Inputs{
		Resources: []v1alpha1.TaskResource{{
			Name: "gitspace",
			Type: "git",
		}},
	}
)

func (r *ReconcileComponent) buildTaskS2iBuildahPush(c *v1alpha2.Component) (*v1alpha1.Task, error) {
	task := &v1alpha1.Task{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "tekton.dev/v1alpha1",
			Kind:       "Task",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: c.Namespace,
			Name:      taskS2iBuildahPusName,
		},
		Spec: v1alpha1.TaskSpec{
			Inputs: &v1alpha1.Inputs{
				Resources: []v1alpha1.TaskResource{{
					Name:       "workspace-git",
					Type:       "git",
					TargetPath: "/",
				}},
				Params: []v1alpha1.TaskParam{
					{Name: "verifyTLS", Default: "true", Description: "Verify registry certificates"},
					{Name: "contextFolder", Default: ".", Description: "the path of the context to build"},
					{Name: "baseImage", Description: "s2i base image"},
				}},
			Outputs: &v1alpha1.Outputs{
				Resources: []v1alpha1.TaskResource{{
					Name: "image",
					Type: "image",
				}},
			},
			Steps: []corev1.Container{
				// # Generate a Dockerfile using the s2i tool
				{
					Name:  "generate",
					Image: "quay.io/openshift-pipeline/s2i-buildah:latest",
					Args: []string{
						"${inputs.params.contextFolder}",
						"${inputs.params.baseImage}",
						"${outputs.resources.image.url}",
						"--image-scripts-url",
						"image:///usr/local/s2i"},
					WorkingDir:   "/sources",
					VolumeMounts: []corev1.VolumeMount{
						{
							MountPath: "/sources",
							Name: "generatedsources"},
					    },
				},
				// Build a Container image using the dockerfile created previously
				{
					Name:    "build",
					Image:   "quay.io/openshift-pipeline/buildah:testing",
					Command: []string{
						"buildah",
					},
					Args: []string{
						"bud",
						"--layers",
						"--tls-verify=${inputs.params.verifyTLS}",
						"-f",
						"Dockerfile",
						"-t",
						"${outputs.resources.image.url}",
						"/sources"},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name: "libcontainers",
							MountPath: "/var/lib/containers",

						},
						{
							Name: "generatedsources",
							MountPath: "/sources",
						},
					},
					SecurityContext: &corev1.SecurityContext{
						Privileged: newBoolPtr(true),
					},
				},
				// Push the image created to quay.io using as credentials the secret mounted within
				// the service account
				{
					Name:    "push",
					Image:   "quay.io/openshift-pipeline/buildah:testing",
					Command: []string{
						"buildah",
					},
					Args:    []string{
						"push",
						"--layers",
						"--tls-verify=${inputs.params.verifyTLS}",
						"${outputs.resources.image.url}",
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							MountPath: "/var/lib/containers",
							Name: "libcontainers"},
					},
					SecurityContext: &corev1.SecurityContext{
					 	Privileged: newBoolPtr(true),
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "generatedsources",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
				{
					Name: "libcontainers",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
			},
		},
	}

	// Set Component instance as the owner and controller
	controllerutil.SetControllerReference(c, task, r.scheme)
	return task, nil
}

func newBoolPtr(b bool) *bool {
	return &b
}
