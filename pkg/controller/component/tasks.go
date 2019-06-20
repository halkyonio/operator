package component

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	v1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	taskS2iBuildahPushName = "s2i-buildah-push"
	serviceAccountName     = "build-bot"
)

func (r *ReconcileComponent) buildTaskS2iBuildahPush(res dependentResource, c *v1alpha2.Component) (runtime.Object, error) {
	task := &v1alpha1.Task{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "tekton.dev/v1alpha1",
			Kind:       "Task",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: c.Namespace,
			Name:      res.name(c),
		},
		Spec: v1alpha1.TaskSpec{
			Inputs: &v1alpha1.Inputs{
				// This input corresponds to a pre-task step as the project will be cloned
				// under by default the following directory : /workspace/{resource-name}
				// Resource name has been defined to git hereafter
				Resources: []v1alpha1.TaskResource{{
					Name:       "git",
					Type:       "git",
				}},
				Params: []v1alpha1.TaskParam{
					{Name: "baseImage", Default: "quay.io/snowdrop/spring-boot-maven-s2i", Description: "S2i base image"},
					{Name: "contextPath", Default:".", Description:"The location of the path to run s2i from"},
					{Name: "moduleDirName", Default: ".", Description: "The name of the directory containing the project (maven, ...) to be compiled"},
					{Name: "verifyTLS", Default: "false", Description: "Verify registry certificates"},
					{Name: "workspacePath", Default: "/workspace/git",Description: "Git path where project is cloned"},
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
					Image: "quay.io/openshift-pipeline/s2i-buildah",
					Args: []string{
						"${inputs.params.contextPath}",
						"${inputs.params.baseImage}",
						"--as-dockerfile",
						"/sources/Dockerfile.gen",
						"--image-scripts-url",
						"image:///usr/local/s2i",
						"--env",
						"MAVEN_ARGS_APPEND=-pl ${inputs.params.moduleDirName}",
						"--env",
						"MAVEN_S2I_ARTIFACT_DIRS=${inputs.params.moduleDirName}/target",
						"--env",
						"S2I_SOURCE_DEPLOYMENTS_FILTER=*.jar",
					},
					WorkingDir: "${inputs.params.workspacePath}",
					VolumeMounts: []corev1.VolumeMount{
						{
							MountPath: "/sources",
							Name:      "generatedsources"},
					},
				},
				// Build a Container image using the dockerfile created previously
				{
					Name:  "build",
					Image: "quay.io/openshift-pipeline/buildah",
					WorkingDir: "/sources",
					Command: []string{
						"buildah",
					},
					Args: []string{
						"bud",
						"--tls-verify=${inputs.params.verifyTLS}",
						"--layers",
						"-f",
						"/sources/Dockerfile.gen",
						"-t",
						"${outputs.resources.image.url}",
						"."},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "libcontainers",
							MountPath: "/var/lib/containers",
						},
						{
							Name:      "generatedsources",
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
					Name:  "push",
					Image: "quay.io/openshift-pipeline/buildah",
					Command: []string{
						"buildah",
					},
					Args: []string{
						"push",
						"--tls-verify=${inputs.params.verifyTLS}",
						"${outputs.resources.image.url}",
						"docker://${outputs.resources.image.url}",
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							MountPath: "/var/lib/containers",
							Name:      "libcontainers"},
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
	controllerutil.SetControllerReference(c, task, r.Scheme)
	return task, nil
}

func newBoolPtr(b bool) *bool {
	return &b
}
