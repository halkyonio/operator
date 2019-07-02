package component

import (
	"context"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Delete an ImageStream created by OpenShift if it exists as we don't own this resource
func (r *ReconcileComponent) deleteImageStream(request reconcile.Request) error {

	imageStream := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "image.openshift.io/v1",
			"kind":       "ImageStream",
			"metadata": map[string]interface{}{
				"name":      request.Name,
				"namespace": request.Namespace,
			},
		},
	}

	e := r.client.Get(context.TODO(), types.NamespacedName{Namespace: request.Namespace, Name: request.Name}, imageStream)
	if e == nil {
		// An imageStream exists, so we will delete it
		e := r.client.Delete(context.TODO(), imageStream)
		if e != nil {
			return e
		}
	} else {
		return e
	}
	return nil
}
