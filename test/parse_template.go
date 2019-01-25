package main

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha1"
	"html/template"
	"os"
)

// In the template, we use rangeStruct to turn our struct values
// into a slice we can iterate over
var k8sTemplate = `apiVersion: apps.openshift.io/v1
kind: DeploymentConfig
metadata:
  labels:
    app: {{.Spec.Name}}{{ range $key, $value := .Labels }}
    {{ $key }}: {{ $value }}{{ end }}
  name: {{.Spec.Name}}
spec:
  replicas: 1
  selector:
    app: {{.Name}}
    deploymentconfig: {{.Name}}
  strategy:
    type: Rolling
  template:
    metadata:
      labels:
        app: {{.Name}}
        deploymentconfig: {{.Name}}
      name: {{.Name}}
    spec:
      containers:
      - args:
        - -c
        - /var/lib/supervisord/conf/supervisor.conf
        command:
        - /var/lib/supervisord/bin/supervisord
        env:
        {{ range .Spec.Envs }}
        - name: {{.Name}}
          value: {{.Value}}
        {{end}}
        image: {{.Spec.RuntimeName}}:latest
        imagePullPolicy: Always
        name: {{.Name}}
        ports:
        - containerPort: {{.Spec.Port}}
          protocol: TCP
        volumeMounts:
        - mountPath: /var/lib/supervisord
          name: shared-data
        - mountPath: /tmp/artifacts
          name: {{.Spec.Storage.Name}}
      initContainers:
      - env:
        - name: CMDS
          value: run-java:/usr/local/s2i/run;run-node:/usr/libexec/s2i/run;compile-java:/usr/local/s2i/assemble;build:/deployments/buildapp
        image: "{{.Spec.SupervisordName}}:latest"
        imagePullPolicy: Always
        name: copy-supervisord
        resources: {}
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /var/lib/supervisord
          name: shared-data
      volumes:
      - emptyDir: {}
        name: shared-data
      - name: {{.Spec.Storage.Name}}
        persistentVolumeClaim:
          claimName: {{.Spec.Storage.Name}}
  triggers:
  - type: ImageChange
    imageChangeParams:
      automatic: true
      containerNames:
      - copy-supervisord
      from:
        kind: ImageStreamTag
        name: copy-supervisord:latest
  - type: ImageChange
    imageChangeParams:
      automatic: true
      containerNames:
      - {{.Name}}
      from:
        kind: ImageStreamTag
        name: {{.Spec.RuntimeName}}:latest`

func main() {

	component := &v1alpha1.Component{
		Spec: v1alpha1.ComponentSpec{
			Name: "dabou",
			Runtime: "spring-boot",
		},
	}
	EnrichComponentWithLabels(component)

	t := template.New("t")
	t, err := t.Parse(k8sTemplate)
	if err != nil {
		panic(err)
	}

	err = t.Execute(os.Stdout, component)
	if err != nil {
		panic(err)
	}
}

func EnrichComponentWithLabels(component *v1alpha1.Component) {
	labels := map[string]string{}
	labels[v1alpha1.ComponentLabelKey] = component.Spec.Runtime
	labels[v1alpha1.NameLabelKey] = component.Spec.Name
	labels[v1alpha1.ManagedByLabelKey] = "component-operator"
	component.Labels = labels
}
