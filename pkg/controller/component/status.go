package component

import (
	"context"
	"fmt"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"reflect"
)

func (r *ReconcileComponent) updateStatusInstance(status v1alpha2.Phase, instance *v1alpha2.Component) error {
	r.reqLogger.Info("Updating App Status for the MobileSecurityService")
	instance.Status.Phase = status
	if err := r.client.Update(context.TODO(), instance); err != nil && k8serrors.IsConflict(err) {
		log.Info("Component status update failed")
		return err
	}
	r.reqLogger.Info(fmt.Sprintf("Status updated : %s", instance.Status.Phase))
	return nil
}

//updateStatus
func (r *ReconcileComponent) updateStatus(podStatus *corev1.Pod, instance *v1alpha2.Component) error {
	r.reqLogger.Info("Updating Component status")
	var status v1alpha2.Phase
	if podStatus != nil  {
		// Pod status is Ready
		status = v1alpha2.PhaseReady
	} else {
		status = v1alpha2.PhaseDeploying
	}

	if !reflect.DeepEqual(status, instance.Status.Phase) {
		instance.Status.Phase = status
		//err := r.client.Status().Update(context.TODO(), instance)
		err := r.client.Update(context.TODO(),instance)
		if err != nil {
			r.reqLogger.Error(err, "Failed to update Status for the Component")
			return err
		}
	}
	return nil
}

// Check if the Pod matching the label selector of our Component contains as condition
// Type = Ready and Status = True
func (r *ReconcileComponent) checkPodReady(instance *v1alpha2.Component) (*corev1.Pod, error) {
	pods, err := r.fetchPod(instance)
	if err != nil {
		return nil, err
	}
	if len(pods.Items) > 0 {
		pod := pods.Items[0]
		for _, c := range pod.Status.Conditions {
			if c.Type == corev1.PodReady && c.Status == corev1.ConditionTrue {
				return &pod, nil
			}
		}
	}

	return nil, nil
}
