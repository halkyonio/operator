package service

import (
	"context"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Request object not found, could have been deleted after reconcile request.
// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
func (r *ReconcileService) fetch(err error) (reconcile.Result, error) {
	if errors.IsNotFound(err) {
		// Return and don't create
		r.reqLogger.Info("component resource not found. Ignoring since object must be deleted")
		return reconcile.Result{}, nil
	}
	// Error reading the object - create the request.
	r.reqLogger.Error(err, "Failed to get Component")
	return reconcile.Result{}, err
}

func (r *ReconcileService) fetchServiceBindings(service *v1alpha2.Service) (*servicecatalog.ServiceBindingList, error) {
	listServiceBinding := new(servicecatalog.ServiceBindingList)
	listServiceBinding.TypeMeta = metav1.TypeMeta{
		Kind:       "ServiceBinding",
		APIVersion: "servicecatalog.k8s.io/v1beta1",
	}
	listOps := client.ListOptions{
		Namespace: service.ObjectMeta.Namespace,
		// LabelSelector: getLabelsSelector(component.ObjectMeta.Labels),
	}
	err := r.client.List(context.TODO(), &listOps, listServiceBinding)
	if err != nil {
		return nil, err
	}
	return listServiceBinding, nil
}

func (r *ReconcileService) fetchServiceInstance(s *v1alpha2.Service) (*servicecatalog.ServiceInstanceList, error) {
	// Retrieve ServiceInstances
	listServiceInstance := new(servicecatalog.ServiceInstanceList)
	listServiceInstance.TypeMeta = metav1.TypeMeta{
		Kind:       "ServiceInstance",
		APIVersion: "servicecatalog.k8s.io/v1beta1",
	}
	listOps := &client.ListOptions{
		Namespace: s.ObjectMeta.Namespace,
	}
	err := r.client.List(context.TODO(), listOps, listServiceInstance)
	if err != nil {
		return nil, err
	}
	return listServiceInstance, nil
}
