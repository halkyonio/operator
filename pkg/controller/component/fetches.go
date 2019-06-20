package component

import (
	"context"
	"fmt"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Request object not found, could have been deleted after reconcile request.
// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
func (r *ReconcileComponent) fetch(err error) (reconcile.Result, error) {
	if errors.IsNotFound(err) {
		// Return and don't create
		r.reqLogger.Info("component resource not found. Ignoring since object must be deleted")
		return reconcile.Result{}, nil
	}
	// Error reading the object - create the request.
	r.reqLogger.Error(err, "Failed to get Component")
	return reconcile.Result{}, err
}

func (r *ReconcileComponent) fetchComponent(request reconcile.Request) (*v1alpha2.Component, error) {
	component := &v1alpha2.Component{}
	err := r.client.Get(context.TODO(), request.NamespacedName, component)
	return component, err
}

func (r *ReconcileComponent) fetchTaskRun(c *v1alpha2.Component) (*tektonv1alpha1.TaskRun, error) {
	taskRun := &tektonv1alpha1.TaskRun{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: c.Name, Namespace: c.Namespace}, taskRun)
	return taskRun, err
}

//fetchPod returns the pod resource created for this instance and where label app=component name
func (r *ReconcileComponent) fetchPod(instance *v1alpha2.Component) (*corev1.Pod, error) {
	pods := &corev1.PodList{}
	lo := &client.ListOptions{}
	lo.InNamespace(instance.Namespace)
	lo.MatchingLabels(map[string]string{"app": instance.Name})
	if err := r.client.List(context.TODO(), lo, pods); err != nil {
		r.reqLogger.Info("Pod(s) don't exist")
		return &corev1.Pod{}, err
	} else {
		// We assume that there is only one Pod containing the label app=component name AND we return it
		if len(pods.Items) > 0 {
			return &pods.Items[0], nil
		} else {
			err := fmt.Errorf("failed to get pod created for the component")
			return &corev1.Pod{}, err
		}
	}
}
