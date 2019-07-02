package capability

import (
	"context"
	kubedbv1 "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"k8s.io/apimachinery/pkg/types"
)

// Request object not found, could have been deleted after reconcile request.
// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
/*func (r *ReconcileCapability) fetch(err error) (reconcile.Result, error) {
	if errors.IsNotFound(err) {
		// Return and don't create
		r.reqLogger.Info("component resource not found. Ignoring since object must be deleted")
		return reconcile.Result{}, nil
	}
	// Error reading the object - create the request.
	r.reqLogger.Error(err, "Failed to get Component")
	return reconcile.Result{}, err
}
*/

func (r *ReconcileCapability) fetchKubeDBPostgres(c *v1alpha2.Capability) (*kubedbv1.Postgres, error) {
	// Retrieve Postgres DB CRD
	postgres := &kubedbv1.Postgres{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Namespace: c.Namespace, Name: c.Name}, postgres)
	return postgres, err
}
