package capability

import (
	"errors"
	"fmt"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
)

func newFalse() *bool {
	b := false
	return &b
}

func (r *ReconcileCapability) installDB(c *v1alpha2.Capability) (bool, error) {

	hasChanges := newFalse()
	// Check if the Secret exists
	if _, e := r.fetchSecret(c); e != nil {
		if e = r.create(c, SECRET); e != nil {
			return false, e
		} else {
			*hasChanges = true
		}
	}

	if string(c.Spec.Kind) == string(v1alpha2.PostgresKind) {
		// Check if the KubeDB - Postgres exists
		if _, e := r.fetchKubeDBPostgres(c); e != nil {
			if e = r.create(c, PG_DATABASE); e != nil {
				return false, e
			} else {
				*hasChanges = true
			}
		}
	} else {
		return false, fmt.Errorf("Database kind not supported %s", c.Spec.Kind)
	}

	return *hasChanges, nil
}
