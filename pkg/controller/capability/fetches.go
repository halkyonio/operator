package capability

import (
	"context"
	kubedbv1 "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"k8s.io/apimachinery/pkg/types"
)

func (r *ReconcileCapability) fetchKubeDBPostgres(c *v1alpha2.Capability) (*kubedbv1.Postgres, error) {
	// Retrieve Postgres DB CRD
	postgres := &kubedbv1.Postgres{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Namespace: c.Namespace, Name: c.Name}, postgres)
	return postgres, err
}
