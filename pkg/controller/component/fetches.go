package component

import (
	"context"
	"fmt"
	"halkyon.io/operator/pkg/controller"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//fetchPod returns the pod resource created for this instance and where label app=component name
func (r *ReconcileComponent) fetchPod(instance *controller.Component) (*corev1.Pod, error) {
	pods := &corev1.PodList{}
	lo := &client.ListOptions{}
	lo.InNamespace(instance.Namespace)
	lo.MatchingLabels(map[string]string{"app": instance.Name})
	if err := r.Client.List(context.TODO(), lo, pods); err != nil {
		r.ReqLogger.Info("Pod(s) don't exist")
		return nil, err
	} else {
		// We assume that there is only one Pod containing the label app=component name AND we return it
		if len(pods.Items) > 0 {
			return &pods.Items[0], nil
		} else {
			err := fmt.Errorf("failed to get pod created for the component")
			return nil, err
		}
	}
}
