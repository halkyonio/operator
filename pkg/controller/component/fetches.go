package component

import (
	"context"
	"fmt"
	buildv1 "github.com/openshift/api/build/v1"
	deploymentconfigv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func (r *ReconcileComponent) fetchComponent(request reconcile.Request) (*v1alpha2.Component, error){
	component := &v1alpha2.Component{}
	err := r.client.Get(context.TODO(), request.NamespacedName, component)
	return component, err
}

//fetchRoute returns the Route resource created for this instance
func (r *ReconcileComponent) fetchRoute(instance *v1alpha2.Component) (*routev1.Route, error) {
	route := &routev1.Route{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, route); err != nil {
		r.reqLogger.Info("Route don't exist")
		return route, err
	} else {
		return route, nil
	}
}

//fetchPod returns the pod resource created for this instance and where label app=component name
func (r *ReconcileComponent) fetchPod(instance *v1alpha2.Component) (*corev1.Pod, error) {
	pods := &corev1.PodList{}
	lo := &client.ListOptions{}
	lo.InNamespace(instance.Namespace)
	lo.MatchingLabels(map[string]string{"app": instance.Name})
	if err := r.client.List(context.TODO(), lo, pods); err != nil {
		r.reqLogger.Info("Pod(s) don't exist")
		return &corev1.Pod{}, err
	} else {
		// We assume that there is only one Pod containing the label app=component anem AND we return it
		if (len(pods.Items) > 0) {
			return &pods.Items[0], nil
		} else {
			err := fmt.Errorf("Failed to get The Pod created for the Component")
			return &corev1.Pod{}, err
		}
	}
}

//fetchService returns the service resource created for this instance
func (r *ReconcileComponent) fetchService(instance *v1alpha2.Component) (*corev1.Service, error) {
	service := &corev1.Service{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, service); err != nil {
		r.reqLogger.Info("Service don't exist")
		return service, err
	} else {
		return service, nil
	}
}

//fetchImageStream returns the image stream resources created for this instance
func (r *ReconcileComponent) fetchImageStream(instance *v1alpha2.Component, imageName string) (*imagev1.ImageStream, error) {
	is := &imagev1.ImageStream{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: imageName, Namespace: instance.Namespace}, is); err != nil {
		r.reqLogger.Info("Imagestream don't exist")
		return is, err
	} else {
		return is, nil
	}
}

//fetchImageStreamList returns the image stream resources created for this instance
func (r *ReconcileComponent) fetchImageStreamList(instance *v1alpha2.Component) (*imagev1.ImageStreamList, error) {
	l := &imagev1.ImageStreamList{}
	lo := &client.ListOptions{}
	lo.InNamespace(instance.Namespace)
	lo.MatchingLabels(map[string]string{"app": instance.Name})
	if err := r.client.List(context.TODO(), lo, l); err != nil {
		r.reqLogger.Info("Imagestream don't exist")
		return l, err
	} else {
		return l, nil
	}
}

//fetchDeployment returns the deployment resource created for this instance
func (r *ReconcileComponent) fetchDeployment(instance *v1alpha2.Component) (*v1beta1.Deployment, error) {
	deployment := &v1beta1.Deployment{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, deployment); err != nil {
		r.reqLogger.Info("Deployment don't exist")
		return deployment, err
	} else {
		return deployment, nil
	}
}

//fetchDeploymentConfig returns the deployment config resource created for this instance
func (r *ReconcileComponent) fetchDeploymentConfig(instance *v1alpha2.Component) (*deploymentconfigv1.DeploymentConfig, error) {
	deployment := &deploymentconfigv1.DeploymentConfig{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, deployment); err != nil {
		r.reqLogger.Info("DeploymentConfig don't exist")
		return deployment, err
	} else {
		return deployment, nil
	}
}

//fetchBuildConfig returns the build config resource created for this instance
func (r *ReconcileComponent) fetchBuildConfig(instance *v1alpha2.Component) (*buildv1.BuildConfig, error) {
	build := &buildv1.BuildConfig{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, build); err != nil {
		r.reqLogger.Info("BuildConfig don't exist")
		return build, err
	} else {
		return build, nil
	}
}

//fetchPVC returns the PVC resource created for this instance
func (r *ReconcileComponent) fetchPVC(instance *v1alpha2.Component) (*corev1.PersistentVolumeClaim, error) {
	pvc := &corev1.PersistentVolumeClaim{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: "m2-data-" + instance.Name, Namespace: instance.Namespace}, pvc); err != nil {
		r.reqLogger.Info("PVC don't exist")
		return pvc, err
	} else {
		return pvc, nil
	}
}