package component

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	buildv1 "github.com/openshift/api/build/v1"
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
func (r *ReconcileComponent) buildBuildConfig(c *v1alpha2.Component) *buildv1.BuildConfig {
	_ = r.getAppLabels(c.Name)
	build := &buildv1.BuildConfig{

	}
	// Set Component instance as the owner and controller
	controllerutil.SetControllerReference(c, build, r.scheme)
	return build
}
