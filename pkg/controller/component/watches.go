package component

import (
	buildv1 "github.com/openshift/api/build/v1"
	deploymentconfigv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

//Watch for changes to secondary resources and create the owner MobileSecurityService
//Watch Deployment resources created in the project/namespace
func watchDeployment(c controller.Controller) error {
	err := c.Watch(&source.Kind{Type: &v1beta1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha2.Component{},
	})
	return err
}

//Watch for changes to secondary resources and create the owner MobileSecurityService
//Watch Deployment Config resources created in the project/namespace
func watchDeploymentConfig(c controller.Controller) error {
	err := c.Watch(&source.Kind{Type: &deploymentconfigv1.DeploymentConfig{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha2.Component{},
	})
	return err
}

//Watch for changes to secondary resources and create the owner MobileSecurityService
//Watch Build Config resources created in the project/namespace
func watchBuildConfig(c controller.Controller) error {
	err := c.Watch(&source.Kind{Type: &buildv1.BuildConfig{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha2.Component{},
	})
	return err
}

//Watch for changes to secondary resources and requeue the owner Component
//Watch Capability resources created in the project/namespace
func watchService(c controller.Controller) error {
	err := c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha2.Component{},
	})
	return err
}

//Watch for changes to secondary resources and requeue the owner Component
//Watch Pod resources created in the project/namespace
func watchPod(c controller.Controller) error {
	err := c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha2.Component{},
	})
	return err
}

//Watch for changes to secondary resources and requeue the owner Component
//Watch Route resources created in the project/namespace
func watchRoute(c controller.Controller) error {
	err := c.Watch(&source.Kind{Type: &routev1.Route{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha2.Component{},
	})
	return err
}
