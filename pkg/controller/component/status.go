package component

import (
	"context"
	"fmt"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	// k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"reflect"
)

//updateStatus
func (r *ReconcileComponent) updateStatus(podStatus *corev1.Pod, instance *v1alpha2.Component, request reconcile.Request) error {
	if !r.isPodReady(podStatus) {
		// err := fmt.Errorf("Failed to get Status = Ready for Pod created by the Component")
		// r.reqLogger.Error(err, "One of the resources such as Pod is not yet ready")
		r.reqLogger.Info("One of the resources such as Pod is not yet ready/running. Component status will not been updated yet")
		return nil
	}

	status := v1alpha2.PhaseComponentReady
	if !reflect.DeepEqual(status, instance.Status.Phase) {
		// Get a more recent version of the CR
		component, err := r.fetchComponent(request)
		if err != nil {
			r.reqLogger.Error(err, "Failed to get the Component")
			return err
		}

		component.Status.Phase = status
		// Update the CR

		err = r.client.Status().Update(context.TODO(), component)
		if err != nil {
			r.reqLogger.Error(err, "Failed to update Status of the Component")
			return err
		}
		r.reqLogger.Info("Updating Component status to status Ready")
	}
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

		err = r.client.Status().Update(context.TODO(), component)
		if err != nil {
			r.reqLogger.Error(err, "Failed to update Status of the Component")
			return err
		}
	}
	r.reqLogger.Info(fmt.Sprintf("Updating Component status to status %v", phase))
	return nil
}

//updatePodStatus returns an error when when the Pod resource could not be updated
func (r *ReconcileComponent) updatePodStatus(instance *v1alpha2.Component, request reconcile.Request) (*corev1.Pod, error) {
	r.reqLogger.Info("Updating pod Status for the Component")
	podStatus, err := r.fetchPod(instance)
	if err != nil {
		r.reqLogger.Info("No pod already exists for", "Component.Namespace", instance.Namespace, "Component.Name", instance.Name)
		return podStatus, nil
	}

	if !reflect.DeepEqual(podStatus.Name, instance.Status.PodName) || !reflect.DeepEqual(podStatus.Status, instance.Status.PodStatus) {
		// Get a more recent version of the CR
		component, err := r.fetchComponent(request)
		if err != nil {
			r.reqLogger.Error(err, "Failed to get the Component")
			return podStatus, err
		}

		component.Status.PodName = podStatus.Name
		component.Status.PodStatus = podStatus.Status

		err = r.client.Status().Update(context.TODO(), component)
		if err != nil {
			r.reqLogger.Error(err, "Failed to update Pod Name and Pod Status for the Component")
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
