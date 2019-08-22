package capability

import (
	"fmt"
	kubedbv1 "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	authorizv1 "github.com/openshift/api/authorization/v1"
	capability "halkyon.io/api/capability/v1beta1"
	v1 "k8s.io/api/core/v1"
	"strings"
)

func (r *ReconcileCapability) installDB(c *capability.Capability) (e error) {
	if e = r.CreateIfNeeded(c, &authorizv1.Role{}); e != nil {
		return e
	}
	if e = r.CreateIfNeeded(c, &authorizv1.RoleBinding{}); e != nil {
		return e
	}

	if e = r.CreateIfNeeded(c, &v1.Secret{}); e != nil {
		return e
	}

	if string(c.Spec.Type) == strings.ToLower(string(capability.PostgresType)) {
		// Check if the KubeDB - Postgres exists
		if e = r.CreateIfNeeded(c, &kubedbv1.Postgres{}); e != nil {
			return e
		}
	} else {
		return fmt.Errorf("unsupported '%s' database kind", c.Spec.Type)
	}

	return
}
