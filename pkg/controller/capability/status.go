package capability

import (
	"context"
	servicecatalogv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)


//updateStatus returns error when status regards the all required resources could not be updated
func (r *ReconcileCapability) updateStatus(serviceBindingStatus *servicecatalogv1beta1.ServiceBinding, serviceInstanceStatus *servicecatalogv1beta1.ServiceInstance, instance *v1alpha2.Capability, request reconcile.Request) error {
	if r.isServiceBindingReady(serviceBindingStatus) && r.isServiceInstanceReady(serviceInstanceStatus) {
		r.reqLogger.Info("Updating Status of the Capability to Ready")
		status := v1alpha2.PhaseComponentReady
		if !reflect.DeepEqual(status, instance.Status.Phase) {
			// Get a more recent version of the CR
			service, err := r.fetchService(request)
			if err != nil {
				r.reqLogger.Error(err, "Failed to get the Capability")
				return err
			}

			service.Status.Phase = v1alpha2.PhaseCapabilityReady
			err = r.client.Status().Update(context.TODO(), service)
			if err != nil {
				r.reqLogger.Error(err, "Failed to update Status for the Capability App")
				return err
			}
		}
		return nil
	} else {
		r.reqLogger.Info("Capability instance or binding are not yet ready. So, we won't update the status of the Capability to Ready", "Namespace", instance.Namespace, "Name", instance.Name)
		return nil
	}
}

//updateStatus
func (r *ReconcileCapability) updateServiceStatus(instance *v1alpha2.Capability, phase v1alpha2.Phase, request reconcile.Request) error {
	if !reflect.DeepEqual(phase, instance.Status.Phase) {
		// Get a more recent version of the CR
		service, err := r.fetchService(request)
		if err != nil {
			return err
		}

		service.Status.Phase = phase

		err = r.client.Status().Update(context.TODO(), service)
		if err != nil {
			r.reqLogger.Error(err, "Failed to update Status of the Capability")
			return err
		}
	}
	r.reqLogger.Info("Updating Capability status to status Ready")
	return nil
}

//updateServiceBindingStatus returns error when status regards the Capability Binding resource could not be updated
func (r *ReconcileCapability) updateServiceBindingStatus(instance *v1alpha2.Capability, request reconcile.Request) (*servicecatalogv1beta1.ServiceBinding, error) {
	r.reqLogger.Info("Updating ServiceBinding Status for the Capability")
	serviceBinding, err := r.fetchServiceBinding(instance)
	if err != nil {
		r.reqLogger.Error(err, "Failed to get ServiceBinding for Status", "Namespace", instance.Namespace, "Name", instance.Name)
		return serviceBinding, err
	}
	if !reflect.DeepEqual(serviceBinding.Name, instance.Status.ServiceBindingName) || !reflect.DeepEqual(serviceBinding.Status, instance.Status.ServiceBindingStatus) {
		// Get a more recent version of the CR
		service, err := r.fetchService(request)
		if err != nil {
			r.reqLogger.Error(err, "Failed to get the Capability")
			return serviceBinding, err
		}

		service.Status.ServiceBindingName = serviceBinding.Name
		service.Status.ServiceBindingStatus = serviceBinding.Status

		err = r.client.Status().Update(context.TODO(), service)
		if err != nil {
			r.reqLogger.Error(err, "Failed to update ServiceBinding Name and Status for the Capability")
			return serviceBinding, err
		}
		r.reqLogger.Info("ServiceBinding Status updated for the Capability")
	}
	return serviceBinding, nil
}

//updateServiceInstanceStatus returns error when status regards the Capability Instance resource could not be updated
func (r *ReconcileCapability) updateServiceInstanceStatus(instance *v1alpha2.Capability, request reconcile.Request) (*servicecatalogv1beta1.ServiceInstance, error) {
	r.reqLogger.Info("Updating ServiceInstance Status for the Capability")
	serviceInstance, err := r.fetchServiceInstance(instance)
	if err != nil {
		r.reqLogger.Error(err, "Failed to get Capability Instance for Status", "Namespace", instance.Namespace, "Name", instance.Name)
		return serviceInstance, err
	}
	if !reflect.DeepEqual(serviceInstance.Name, instance.Status.ServiceInstanceName) || !reflect.DeepEqual(serviceInstance.Status, instance.Status.ServiceInstanceStatus) {
		// Get a more recent version of the CR
		service, err := r.fetchService(request)
		if err != nil {
			r.reqLogger.Error(err, "Failed to get the Component")
			return serviceInstance, err
		}

		service.Status.ServiceInstanceName = serviceInstance.Name
		service.Status.ServiceInstanceStatus = serviceInstance.Status

		err = r.client.Status().Update(context.TODO(), service)
		if err != nil {
			r.reqLogger.Error(err, "Failed to update ServiceInstance Name and Status for the Capability")
			return serviceInstance, err
		}
		r.reqLogger.Info("ServiceInstance Status updated for the Capability")
	}
	return serviceInstance, nil
}

func (r *ReconcileCapability) isServiceInstanceReady(serviceInstanceStatus *servicecatalogv1beta1.ServiceInstance) bool {
	for _, c := range serviceInstanceStatus.Status.Conditions {
		if c.Type == servicecatalogv1beta1.ServiceInstanceConditionReady && c.Status == servicecatalogv1beta1.ConditionTrue {
			return true
		}
	}
	return false
}

func (r *ReconcileCapability) isServiceBindingReady(serviceBindingStatus *servicecatalogv1beta1.ServiceBinding) bool {
	for _, c := range serviceBindingStatus.Status.Conditions {
		if c.Type == servicecatalogv1beta1.ServiceBindingConditionReady && c.Status == servicecatalogv1beta1.ConditionTrue {
			return true
		}
	}
	return false
}


