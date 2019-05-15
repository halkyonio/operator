package component

import (
	"context"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	// k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
)

//updateStatus
func (r *ReconcileComponent) updateStatus(podStatus *corev1.Pod, instance *v1alpha2.Component) error {
	if !r.isPodReady(podStatus) {
		// err := fmt.Errorf("Failed to get Status = Ready for Pod created by the Component")
		// r.reqLogger.Error(err, "One of the resources such as Pod is not yet ready")
		r.reqLogger.Info("One of the resources such as Pod is not yet ready/runnin. Component status will not been updated yet")
		return nil
	}

	status := v1alpha2.PhaseComponentReady
	if !reflect.DeepEqual(status, instance.Status.Phase) {
		// Get a more recent version of the CR
		component := &v1alpha2.Component{}
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, component)
		if err != nil {
			r.reqLogger.Error(err, "Failed to get the Component")
			return err
		}

		component.Status.Phase = status
		//err := r.client.Status().Update(context.TODO(), instance)
		err = r.client.Update(context.TODO(), component)
		if err != nil {
			r.reqLogger.Error(err, "Failed to update Status of the Component")
			return err
		}
	}
	r.reqLogger.Info("Updating Component status to status Ready")
	return nil
}

//updateStatus
func (r *ReconcileComponent) updateComponentStatus(instance *v1alpha2.Component, phase v1alpha2.Phase, request reconcile.Request) error {
	if !reflect.DeepEqual(phase, instance.Status.Phase) {
		// Get a more recent version of the CR
		component, err := r.fetchComponent(request)
		if err != nil {
			return err
		}

		component.Status.Phase = phase
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			r.reqLogger.Error(err, "Failed to update Status of the Service")
			return err
		}
	}
	r.reqLogger.Info("Updating Service status to status Ready")
	return nil
}

//updatePodStatus returns an error when when the Pod resource could not be updated
func (r *ReconcileComponent) updatePodStatus(instance *v1alpha2.Component) (*corev1.Pod, error) {
	r.reqLogger.Info("Updating pod Status for the Component")
	podStatus, err := r.fetchPod(instance)
	if err != nil {
		r.reqLogger.Info( "No pod already exists for", "Component.Namespace", instance.Namespace, "Component.Name", instance.Name)
		return podStatus, nil
	}

	// Get a more recent version of the CR
	component := &v1alpha2.Component{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, component)
	if err != nil {
		r.reqLogger.Error(err, "Failed to get the Component")
		return podStatus, err
	}

	if !reflect.DeepEqual(podStatus.Name, instance.Status.PodName) {
		instance.Status.PodName = podStatus.Name
		err := r.client.Update(context.TODO(), instance)
		if err != nil {
			r.reqLogger.Error(err, "Failed to update Pod Name for the Component")
			return podStatus, err
		}
	}
	if !reflect.DeepEqual(podStatus.Status, instance.Status.PodStatus) {
		instance.Status.PodStatus = podStatus.Status
		err := r.client.Update(context.TODO(), instance)
		if err != nil {
			r.reqLogger.Error(err, "Failed to update Pod Status for the Component")
			return podStatus, err
		}
	}
	return podStatus, nil
}


// Check if the Pod Condition is Type = Ready and Status = True
func (r *ReconcileComponent) isPodReady(pod *corev1.Pod) bool {
	for _, c := range pod.Status.Conditions {
		if c.Type == corev1.PodReady && c.Status == corev1.ConditionTrue {
			return true
		}
	}

	return false
}
