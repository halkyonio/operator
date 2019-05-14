package service

import (
	"context"
	"fmt"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
)


//updateStatus returns error when status regards the all required resources could not be updated
func (r *ReconcileService) updateStatus(serviceBindingStatus *servicecatalog.ServiceBinding, serviceInstanceStatus *servicecatalog.ServiceInstance, instance *v1alpha2.Service) error {
	r.reqLogger.Info("Updating App Status for the Service")
	if len(serviceBindingStatus.UID) < 1 && len(serviceInstanceStatus.Name) < 1 {
		err := fmt.Errorf("Failed to get OK Status for Service")
		r.reqLogger.Error(err, "One of the resources are not created", "Namespace", instance.Namespace, "Name", instance.Name)
		return err
	}
	status:= "OK"
	if !reflect.DeepEqual(status, instance.Status.Phase) {
		instance.Status.Phase = v1alpha2.PhaseServiceReady
		err := r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			r.reqLogger.Error(err, "Failed to update Status for the Service App")
			return err
		}
	}
	return nil
}

//updateStatus
func (r *ReconcileService) updateServiceStatus(instance *v1alpha2.Service, phase v1alpha2.Phase) error {
	if !reflect.DeepEqual(phase, instance.Status.Phase) {
		// Get a more recent version of the CR
		service := &v1alpha2.Service{}
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, instance)
		if err != nil {
			r.reqLogger.Error(err, "Failed to get the Service")
			return err
		}

		service.Status.Phase = phase
		//err := r.client.Status().Update(context.TODO(), instance)
		err = r.client.Update(context.TODO(), service)
		if err != nil {
			r.reqLogger.Error(err, "Failed to update Status of the Service")
			return err
		}
	}
	r.reqLogger.Info("Updating Service status to status Ready")
	return nil
}

//updateServiceBindingStatus returns error when status regards the Service Binding resource could not be updated
func (r *ReconcileService) updateServiceBindingStatus(instance *v1alpha2.Service) (*servicecatalog.ServiceBinding, error) {
	r.reqLogger.Info("Updating ServiceBinding Status for the Service")
	serviceBinding, err := r.fetchServiceBinding(instance)
	if err != nil {
		r.reqLogger.Error(err, "Failed to get ServiceBinding for Status", "Namespace", instance.Namespace, "Name", instance.Name)
		return serviceBinding, err
	}
	if !reflect.DeepEqual(serviceBinding.Name, instance.Status.ServiceBindingName) {
		instance.Status.ServiceBindingName = serviceBinding.Name
		err := r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			r.reqLogger.Error(err, "Failed to update ServiceBinding Name Status for the Service")
			return serviceBinding, err
		}
	}
	if !reflect.DeepEqual(serviceBinding.Status, instance.Status.ServiceBindingStatus) {
		instance.Status.ServiceBindingStatus = serviceBinding.Status
		err := r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			r.reqLogger.Error(err, "Failed to update ServiceBinding Status for the Service")
			return serviceBinding, err
		}
	}
	return serviceBinding, nil
}

//updateServiceInstanceStatus returns error when status regards the Service Instance resource could not be updated
func (r *ReconcileService) updateServiceInstanceStatus(instance *v1alpha2.Service) (*servicecatalog.ServiceInstance, error) {
	r.reqLogger.Info("Updating Service Instance Status for the Service")
	serviceInstance, err := r.fetchServiceInstance(instance)
	if err != nil {
		r.reqLogger.Error(err, "Failed to get Service Instance for Status", "Namespace", instance.Namespace, "Name", instance.Name)
		return serviceInstance, err
	}
	if !reflect.DeepEqual(serviceInstance.Name, instance.Status.ServiceInstanceName) {
		instance.Status.ServiceInstanceName = serviceInstance.Name
		err := r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			r.reqLogger.Error(err, "Failed to update Service Instance Name Status for the Service")
			return serviceInstance, err
		}
	}
	if !reflect.DeepEqual(serviceInstance.Status, instance.Status.ServiceInstanceStatus) {
		instance.Status.ServiceInstanceStatus = serviceInstance.Status
		err := r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			r.reqLogger.Error(err, "Failed to update Service Instance Status for the Service")
			return serviceInstance, err
		}
	}
	return serviceInstance, nil
}


