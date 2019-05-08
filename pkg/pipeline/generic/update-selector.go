package generic

import (
 	log "github.com/sirupsen/logrus"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"github.com/snowdrop/component-operator/pkg/pipeline"
	"golang.org/x/net/context"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/rest"

	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewUpdateServiceSelectorStep() pipeline.Step {
	return &serviceSelectorStep{}
}

type serviceSelectorStep struct{}

func (serviceSelectorStep) Name() string {
	return "update-service-selector"
}

func (serviceSelectorStep) CanHandle(component *v1alpha2.Component) bool {
	// log.Infof("## Status to be checked : %s", component.Status.Phase)
	return true
}

func (serviceSelectorStep) Handle(component *v1alpha2.Component, config *rest.Config, client *client.Client, namespace string, scheme *runtime.Scheme) error {
	return updateSelector(*component, *config, *client, namespace, *scheme)
}

func updateSelector(component v1alpha2.Component, config rest.Config, c client.Client, namespace string, scheme runtime.Scheme) error {
	component.ObjectMeta.Namespace = namespace
	componentName := component.Annotations["app.openshift.io/component-name"]
	svc := &v1.Service{}
	svc.Labels = map[string]string{
		"app.kubernetes.io/name": componentName,
		"app.openshift.io/runtime": component.Spec.Runtime,
	}
	if err := c.Get(context.TODO(), types.NamespacedName{Name: componentName, Namespace: namespace}, svc); err != nil {
		return err
	}

	var nameApp string
	if component.Spec.DeploymentMode == "outerloop" {
		nameApp = componentName + "-build"
	} else {
		nameApp = componentName
	}
	svc.Spec.Selector = map[string]string{
		"app": nameApp,
	}
	if err := c.Update(context.TODO(),svc) ; err != nil {
		return err
	}
	log.Infof("### Updated Service Selector to switch to a different component.")
	log.Info("------------------------------------------------------------------")
	return nil
}