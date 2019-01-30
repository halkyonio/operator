package outerloop

import (
	"context"
	"github.com/openshift/api/apps/v1"
	log "github.com/sirupsen/logrus"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha1"
	"github.com/snowdrop/component-operator/pkg/pipeline"
	"github.com/snowdrop/component-operator/pkg/util/kubernetes"
	openshift "github.com/snowdrop/component-operator/pkg/util/openshift"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
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
	componentName := component.Name

	isOpenshift, err := kubernetes.DetectOpenShift()
	if err != nil {
		return err
	}

	if isOpenshift {
		found, err := openshift.GetDeploymentConfig(namespace, componentName, c)
		if err != nil {
			log.Info("### DeploymentConfig %s not found",componentName)
			return err
		}

		clone := &v1.DeploymentConfig{}
		clone.ObjectMeta = found.ObjectMeta
		clone.Name = component.Name + "-build"
		clone.Spec = found.Spec
		clone.Spec.Template.Spec.InitContainers = []corev1.Container{} // Remove initContainers
		container := clone.Spec.Template.Spec.Containers[0]
		container.Args = []string{} // Remove args of the container
		container.VolumeMounts = []corev1.VolumeMount{} // Remove volume to be mounted
		container.Image = clone.Annotations["app.openshift.io/runtime-image"]// Use the runtime/build image
		clone.Spec.Template.Spec.Volumes = []corev1.Volume{} // Remove volume

		// Create the new DC using cloned info
		err = c.Create(context.TODO(), clone)
		if err != nil && k8serrors.IsConflict(err) {
			// Retry function on conflict
			log.Info("### DeploymentConfig update failed due to conflict!")
			return err
		}
	}
	return nil
}
