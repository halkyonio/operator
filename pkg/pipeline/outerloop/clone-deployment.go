package outerloop

import (
	"bytes"
	"context"
	"fmt"
	deploymentconfig "github.com/openshift/api/apps/v1"
	log "github.com/sirupsen/logrus"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha1"
	"github.com/snowdrop/component-operator/pkg/pipeline"
	"github.com/snowdrop/component-operator/pkg/util/kubernetes"
	"github.com/snowdrop/component-operator/pkg/util/openshift"
	util "github.com/snowdrop/component-operator/pkg/util/template"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"text/template"
)

func NewCloneDeploymentStep() pipeline.Step {
	return &cloneDeploymentStep{}
}

type cloneDeploymentStep struct{}

func (cloneDeploymentStep) Name() string {
	return "clone-deployment"
}

func (cloneDeploymentStep) CanHandle(component *v1alpha1.Component) bool {
	// log.Infof("## Status to be checked : %s", component.Status.Phase)
	return true
}

func (cloneDeploymentStep) Handle(component *v1alpha1.Component, client *client.Client, namespace string, scheme *runtime.Scheme) error {
	return cloneDeploymentLoop(component, *client, namespace, *scheme)
}

func cloneDeploymentLoop(component *v1alpha1.Component, c client.Client, namespace string, scheme runtime.Scheme) error {
	component.ObjectMeta.Namespace = namespace

	isOpenshift, err := kubernetes.DetectOpenShift()
	if err != nil {
		return err
	}

	if isOpenshift {

	/*
	    componentName := component.Name
        found, err := openshift.GetDeploymentConfig(namespace, componentName, c)
		if err != nil {
			log.Info("### DeploymentConfig %s not found",componentName)
			return err
		}

		clone := &v1.DeploymentConfig{}
		clone.ObjectMeta = found.ObjectMeta
		clone.ResourceVersion = "" // Remove Resource Version
		clone.CreationTimestamp = metav1.Time{} // Remove Creation time
		clone.Generation = int64(0)

		clone.Name = component.Name + "-build"
		clone.Spec = found.Spec
		clone.Spec.Selector = map[string]string{
			"deploymentconfig": component.Name + "-build",
		}
		clone.Spec.Template.Name = component.Name + "-build"
		clone.Spec.Template.Spec.InitContainers = []corev1.Container{} // Remove initContainers
		container := clone.Spec.Template.Spec.Containers[0]
		container.Args = []string{} // Remove args of the container
		container.Command = []string{} // Remove command of the container
		container.VolumeMounts = []corev1.VolumeMount{} // Remove volume to be mounted
		container.Image = component.Annotations["app.openshift.io/runtime-image"]// Use the runtime/build image
		container.Name = component.Name + "-build"
		clone.Spec.Template.Spec.Containers[0] = container

		clone.Spec.Template.Spec.Volumes = []corev1.Volume{} // Remove volume*/

		tmpl, ok := util.Templates["outerloop/deploymentconfig"]
		if ok {
			originalcomponentName := component.Name

			// Populate the DC using template
			component.Name = component.Name + "-build"
			r, err := ParseTemplateToRuntimeObject(tmpl,component)
			obj, err := kubernetes.RuntimeObjectFromUnstructured(r)
			if err != nil {
				return err
			}

			// Fetch DC
			dc := obj.(*deploymentconfig.DeploymentConfig)
			found, err := openshift.GetDeploymentConfig(namespace, originalcomponentName, c)
			if err != nil {
				log.Info("### DeploymentConfig not found")
				return err
			}
			containerFound := &found.Spec.Template.Spec.Containers[0]
			container := &dc.Spec.Template.Spec.Containers[0]
			container.Env = containerFound.Env
			container.EnvFrom = containerFound.EnvFrom
			container.Env = UpdateEnv(container.Env)
			dc.Namespace = found.Namespace

			err = c.Create(context.TODO(),dc)
			if err != nil {
				log.Info("### DeploymentConfig build creation failed")
				return err
			}
			log.Infof("### Created Build Deployment Config.")
		}
	}
	log.Info("## Pipeline 'outerloop' ended ##")
	log.Info("------------------------------------------------------")
	return nil
}

func UpdateEnv(envs []v1.EnvVar) []v1.EnvVar {
	newEnvs := []v1.EnvVar{}
	for _, s := range envs {
		if s.Name == "JAVA_APP_JAR" {
			newEnvs = append(newEnvs, v1.EnvVar{Name: s.Name, Value: "fruit-backend-sb-0.0.1-SNAPSHOT.jar"})
		} else {
			newEnvs = append(newEnvs, s)
		}
	}
	return newEnvs
}

func ParseTemplateToRuntimeObject(tmpl template.Template, component *v1alpha1.Component) (*unstructured.Unstructured, error) {
	b := Parse(tmpl, component)
	r, err := kubernetes.PopulateKubernetesObjectFromYaml(b.String())
	if err != nil {
		return nil, err
	}
	return r, nil
}


// Parse the file's template using the Application struct
func Parse(t template.Template, obj *v1alpha1.Component) bytes.Buffer {
	var b bytes.Buffer
	err := t.Execute(&b, obj)
	//fmt.Println(&b, obj)
	if err != nil {
		fmt.Println("There was an error:", err.Error())
	}
	log.Debug("Generated :", b.String())
	return b
}
