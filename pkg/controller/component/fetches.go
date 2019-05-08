package component

import (
	"context"
	buildv1 "github.com/openshift/api/build/v1"
	deploymentconfigv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Request object not found, could have been deleted after reconcile request.
// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
func (r *ReconcileComponent) fetch(err error) (reconcile.Result, error) {
	if errors.IsNotFound(err) {
		// Return and don't create
		r.reqLogger.Info("component resource not found. Ignoring since object must be deleted")
		return reconcile.Result{}, nil
	}
	// Error reading the object - create the request.
	r.reqLogger.Error(err, "Failed to get Component")
	return reconcile.Result{}, err
}

//fetchRoute returns the Route resource created for this instance
func (r *ReconcileComponent) fetchRoute(instance *v1alpha2.Component) (*routev1.Route, error) {
	r.reqLogger.Info("Checking if the route already exists")
	route := &routev1.Route{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, route)
	return route, err
}

//fetchPod returns the pod resource created for this instance
func (r *ReconcileComponent) fetchPod(instance *v1alpha2.Component) (*corev1.Pod, error) {
	r.reqLogger.Info("Checking if the service already exists")
	pod := &corev1.Pod{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, pod)
	return pod, err
}

//fetchService returns the service resource created for this instance
func (r *ReconcileComponent) fetchService(instance *v1alpha2.Component) (*corev1.Service, error) {
	r.reqLogger.Info("Checking if the service already exists")
	service := &corev1.Service{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, service)
	return service, err
}

//fetchImageStream returns the image stream resources created for this instance
func (r *ReconcileComponent) fetchImageStream(instance *v1alpha2.Component, imageName string) (*imagev1.ImageStream, error) {
	r.reqLogger.Info("Checking if the image streams already exists")
	is := &imagev1.ImageStream{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: imageName, Namespace: instance.Namespace}, is)
	return is, err
}

//fetchDeployment returns the deployment resource created for this instance
func (r *ReconcileComponent) fetchDeployment(instance *v1alpha2.Component) (*v1beta1.Deployment, error) {
	r.reqLogger.Info("Checking if the deployment already exists")
	deployment := &v1beta1.Deployment{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, deployment)
	return deployment, err
}

//fetchDeploymentConfig returns the deployment config resource created for this instance
func (r *ReconcileComponent) fetchDeploymentConfig(instance *v1alpha2.Component) (*deploymentconfigv1.DeploymentConfig, error) {
	r.reqLogger.Info("Checking if the deployment already exists")
	deployment := &deploymentconfigv1.DeploymentConfig{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, deployment)
	return deployment, err
}

//fetchBuildConfig returns the build config resource created for this instance
func (r *ReconcileComponent) fetchBuildConfig(instance *v1alpha2.Component) (*buildv1.BuildConfig, error) {
	r.reqLogger.Info("Checking if the deployment already exists")
	build := &buildv1.BuildConfig{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, build)
	return build, err
}

//fetchPVC returns the PVC resource created for this instance
func (r *ReconcileComponent) fetchPVC(instance *v1alpha2.Component) (*corev1.PersistentVolumeClaim, error) {
	r.reqLogger.Info("Checking if the deployment already exists")
	pvc := &corev1.PersistentVolumeClaim{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: "m2-data-" + instance.Name, Namespace: instance.Namespace}, pvc)
	return pvc, err
}