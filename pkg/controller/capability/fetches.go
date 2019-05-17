package capability

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
func (r *ReconcileCapability) fetch(err error) (reconcile.Result, error) {
	if errors.IsNotFound(err) {
		// Return and don't create
		r.reqLogger.Info("component resource not found. Ignoring since object must be deleted")
		return reconcile.Result{}, nil
	}
	// Error reading the object - create the request.
	r.reqLogger.Error(err, "Failed to get Component")
	return reconcile.Result{}, err
}

func (r *ReconcileCapability) fetchCapability(request reconcile.Request) (*v1alpha2.Capability, error){
	cap := &v1alpha2.Capability{}
	err := r.client.Get(context.TODO(), request.NamespacedName, cap)
	return cap, err
}

func (r *ReconcileCapability) fetchServiceBinding(service *v1alpha2.Capability) (*servicecatalogv1.ServiceBinding, error) {
	serviceBinding := &servicecatalogv1.ServiceBinding{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: service.Namespace, Name: service.Spec.Name}, serviceBinding)
	return serviceBinding, err
}

func (r *ReconcileCapability) fetchServiceInstance(service *v1alpha2.Capability) (*servicecatalogv1.ServiceInstance, error) {
	// Retrieve ServiceInstances
	serviceInstance := &servicecatalogv1.ServiceInstance{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: service.Namespace, Name: service.Spec.Name}, serviceInstance)
	return serviceInstance, err
}
