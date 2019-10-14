package capability

import (
	kubedbv1 "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	authorizv1 "github.com/openshift/api/authorization/v1"
	"halkyon.io/operator/pkg/controller"
	v1 "k8s.io/api/core/v1"
)

func (r *ReconcileCapability) installDB(c *controller.Capability) (e error) {
	if e = r.CreateIfNeeded(c, &authorizv1.Role{}); e != nil {
		return e
	}
	if e = r.CreateIfNeeded(c, &authorizv1.RoleBinding{}); e != nil {
		return e
	}

	if e = r.CreateIfNeeded(c, &v1.Secret{}); e != nil {
		return e
	}

	if e = r.CreateIfNeeded(c, &kubedbv1.Postgres{}); e != nil {
		return e
	}
	return
}
