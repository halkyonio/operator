package outerloop

import (
	log "github.com/sirupsen/logrus"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha1"
	"github.com/snowdrop/component-operator/pkg/pipeline"
	"github.com/snowdrop/component-operator/pkg/util/kubernetes"
	util "github.com/snowdrop/component-operator/pkg/util/template"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
/*			clone := &v1alpha1.Component{}
			clone.Name = component.Name + "-build"
			clone.Annotations = map[string]string{
				"app.openshift.io/runtime-image": component.Annotations["app.openshift.io/runtime-image"],
			}
			clone.Labels = component.Labels
			clone.Spec.Envs = component.Spec.Envs
			clone.Spec.Port = component.Spec.Port
			clone.Namespace = component.Namespace*/
			component.Name = component.Name + "-build"
			component.Spec.Envs = append(component.Spec.Envs,v1alpha1.Env{Name: "JAVA_APP_DIR", Value: "fruit-backend-sb-0.0.1-SNAPSHOT.jar"})

			_, err := util.ParseTemplateToRuntimeObject(tmpl,component)
			err = kubernetes.CreateResource(tmpl, component, c, &scheme)
			if err != nil {
				return err
			}
			log.Infof("### Created Build Deployment Config.")
		}
	}
	log.Info("## Pipeline 'outerloop' ended ##")
	log.Info("------------------------------------------------------")
	return nil
}
