package component

import (
	"context"
	"fmt"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileComponent) setErrorStatus(instance *v1alpha2.Component, err error) {
	instance.Status.Phase = v1alpha2.ComponentFailed
	r.updateStatusWithMessage(instance, err.Error())
}

func (r *ReconcileComponent) updateStatusWithMessage(instance *v1alpha2.Component, msg string) {
	// fetch latest version to avoid optimistic lock error
	current, err := r.fetchLatestVersion(instance)
	if err != nil {
		r.reqLogger.Error(err, "failed to fetch latest version of component "+instance.Name)
	}

	current.Status.PodName = instance.Status.PodName
	current.Status.Phase = instance.Status.Phase
	current.Status.Message = msg

	err = r.client.Status().Update(context.TODO(), current)
	if err != nil {
		r.reqLogger.Error(err, "failed to update status for component "+current.Name)
	}
}

func (r *ReconcileComponent) updateStatus(instance *v1alpha2.Component, phase v1alpha2.ComponentPhase) error {
	// Get a more recent version of the CR
	component, err := r.fetchLatestVersion(instance)
	if err != nil {
		return err
	}

	r.reqLogger.Info("updating component status")
	pod, err := r.fetchPod(instance)
	if err != nil || !r.isPodReady(pod) {
		msg := fmt.Sprintf("pod is not ready for component '%s' in namespace '%s'", instance.Name, instance.Namespace)
		r.reqLogger.Info(msg)
		instance.Status.Phase = v1alpha2.ComponentPending
		r.updateStatusWithMessage(instance, msg)
		return nil
	}

	if pod.Name != instance.Status.PodName {
		component.Status.PodName = pod.Name
	}

	if phase != instance.Status.Phase {
		component.Status.Phase = phase
	}

	r.updateStatusWithMessage(component, "")
	return nil
}

func (r *ReconcileComponent) fetchLatestVersion(instance *v1alpha2.Component) (*v1alpha2.Component, error) {
	component, err := r.fetchComponent(reconcile.Request{
		NamespacedName: types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace},
	})
	if err != nil {
		r.reqLogger.Error(err, "failed to get the Component")
		return nil, err
	}
	return component, nil
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
