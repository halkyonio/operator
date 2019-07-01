package capability

import (
	"context"
	kubedbv1 "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Request object not found, could have been deleted after reconcile request.
// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
func (r *ReconcileCapability) fetch(err error) (reconcile.Result, error) {
	if errors.IsNotFound(err) {
		// Return and don't create
		r.reqLogger.Info("component resource not found. Ignoring since object must be deleted")
		return reconcile.Result{}, nil
	}
	// Error reading the object - create the request.
	r.reqLogger.Error(err, "Failed to get Component")
	return reconcile.Result{}, err
}

func (r *ReconcileCapability) fetchCapability(request reconcile.Request) (*v1alpha2.Capability, error) {
	cap := &v1alpha2.Capability{}
	err := r.client.Get(context.TODO(), request.NamespacedName, cap)
	return cap, err
}

func (r *ReconcileCapability) fetchSecret(c *v1alpha2.Capability) (*v1.Secret, error) {
	paramsMap := r.ParametersAsMap(c.Spec.Parameters)
	// Retrieve Secret
	secret := &v1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: c.Namespace, Name: r.SetDefaultSecretNameIfEmpty(c.Name, paramsMap[DB_CONFIG_NAME])}, secret)
	return secret, err
}

func (r *ReconcileCapability) fetchKubeDBPostgres(c *v1alpha2.Capability) (*kubedbv1.Postgres, error) {
	// Retrieve Postgres DB CRD
	postgres := &kubedbv1.Postgres{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: c.Namespace, Name: c.Name}, postgres)
	return postgres, err
}
