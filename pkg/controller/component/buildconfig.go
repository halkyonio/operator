package component

import (
	buildv1 "github.com/openshift/api/build/v1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

/*
apiVersion: build.openshift.io/v1
kind: BuildConfig
metadata:
  labels:
    app: {{.Name}}{{ range $key, $value := .ObjectMeta.Labels }}
    {{ $key }}: {{ $value }}{{ end }}
  name: {{.Name}}
spec:
  output:
    to:
      kind: ImageStreamTag
      name: {{.Name}}:latest
  source:
    git:
      uri: {{ index .ObjectMeta.Annotations "app.openshift.io/git-uri" }}
      ref: {{ index .ObjectMeta.Annotations "app.openshift.io/git-ref" }}
    type: Git
  strategy:
    sourceStrategy:
      from:
        kind: ImageStreamTag
        name: {{.Spec.RuntimeName}}:latest
      incremental: true
      env:
      - name: MAVEN_ARGS_APPEND
        value: "-pl {{ index .ObjectMeta.Annotations "app.openshift.io/git-dir" }}"
      - name: ARTIFACT_DIR
        value: "{{ index .ObjectMeta.Annotations "app.openshift.io/git-dir" }}/target"
      - name: ARTIFACT_COPY_ARGS
        value: "{{ index .ObjectMeta.Annotations "app.openshift.io/artifact-copy-args" }}"
    type: Source
  triggers:
  - github:
      secret: GITHUB_WEBHOOK_SECRET
    type: GitHub
  - type: ConfigChange
  - imageChange: {}
    type: ImageChange
*/

//buildDeployment returns the Deployment config object
func (r *ReconcileComponent) buildBuildConfig(c *v1alpha2.Component) (*buildv1.BuildConfig, error) {
	ls := r.getAppLabels(c.Name)
    buildImage, err:= r.getImageReference(c.Spec)
    if err != nil {
		return nil, err
	}
	build := &buildv1.BuildConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "build.openshift.io/v1",
			Kind:       "BuildConfig",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.Name,
			Namespace: c.Namespace,
			Labels:    ls,
		},
		Spec: buildv1.BuildConfigSpec{
			CommonSpec: buildv1.CommonSpec{
				Output: buildv1.BuildOutput{
					To: &corev1.ObjectReference{
						Kind: "ImageStreamTag",
						// TODO -> Get Image Name of the Microservice
						Name: "TODO"},
				},
				Source: buildv1.BuildSource{
					Type: buildv1.BuildSourceGit,
					Git: &buildv1.GitBuildSource{
						// TODO -> Get k8s annotation
						Ref: "app.openshift.io/git-ref",
						URI: "app.openshift.io/git-uri",
					},
				},
				Strategy: buildv1.BuildStrategy{
					Type: "Source",
					SourceStrategy: &buildv1.SourceBuildStrategy{
						From: corev1.ObjectReference{
							Kind: "DockerImage",
							Name: buildImage},
						Env: []corev1.EnvVar{
							// TODO
							/*
							   - name: MAVEN_ARGS_APPEND
							     value: "-pl {{ index .ObjectMeta.Annotations "app.openshift.io/git-dir" }}"
							   - name: ARTIFACT_DIR
							     value: "{{ index .ObjectMeta.Annotations "app.openshift.io/git-dir" }}/target"
							   - name: ARTIFACT_COPY_ARGS
							     value: "{{ index .ObjectMeta.Annotations "app.openshift.io/artifact-copy-args" }}"
							*/
							{},
							// TODO ->       incremental: true
						},
					},
				}},
			Triggers: []buildv1.BuildTriggerPolicy{
				{Type: buildv1.GitHubWebHookBuildTriggerType,
					GitHubWebHook: &buildv1.WebHookTrigger{Secret: "GITHUB_WEBHOOK_SECRET"},
				},
			},
		},
	}
	// Set Component instance as the owner and controller
	controllerutil.SetControllerReference(c, build, r.scheme)
	return build, nil
}
