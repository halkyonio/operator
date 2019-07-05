package capability

import (
	"fmt"
	kubedbv1 "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	v1 "k8s.io/api/core/v1"
	"strings"
)

func (r *ReconcileCapability) installDB(c *v1alpha2.Capability) (e error) {
	if e = r.CreateIfNeeded(c, &v1.Secret{}); e != nil {
		return e
	}

	if string(c.Spec.Kind) == strings.ToLower(string(v1alpha2.PostgresKind)) {
		// Check if the KubeDB - Postgres exists
		if e = r.CreateIfNeeded(c, &kubedbv1.Postgres{}); e != nil {
			return e
		}
	} else {
		return fmt.Errorf("unsupported '%s' database kind", c.Spec.Kind)
	}

	return
}
