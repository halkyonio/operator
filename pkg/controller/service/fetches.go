package service

import (
	"context"
	servicecatalogv1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
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

func (r *ReconcileService) fetchService(request reconcile.Request) (*v1alpha2.Service, error){
	service := &v1alpha2.Service{}
	err := r.client.Get(context.TODO(), request.NamespacedName, service)
	return service, err
}

func (r *ReconcileService) fetchServiceBinding(service *v1alpha2.Service) (*servicecatalogv1.ServiceBinding, error) {
	serviceBinding := &servicecatalogv1.ServiceBinding{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: service.Namespace, Name: service.Name}, serviceBinding)
	if err != nil {
		return nil, err
	}
	return serviceBinding, nil
}

func (r *ReconcileService) fetchServiceInstance(s *v1alpha2.Service) (*servicecatalogv1.ServiceInstance, error) {
	// Retrieve ServiceInstances
	serviceInstance := &servicecatalogv1.ServiceInstance{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: s.Namespace, Name: s.Name}, serviceInstance)
	if err != nil {
		return nil, err
	}
	return serviceInstance, nil
}
