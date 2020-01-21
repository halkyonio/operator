package component

import (
	"halkyon.io/operator-framework"
	authorizv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type role struct {
	framework.Role
}

func newRole(owner *Component) role {
	return role{Role: framework.NewOwnedRole(owner)}
}

func (res role) Build(empty bool) (runtime.Object, error) {
	ser, err := res.Role.Build(empty)
	if err != nil {
		return nil, err
	}
	r := ser.(*authorizv1.Role)
	if !empty {
		r.Rules = append(r.Rules, authorizv1.PolicyRule{
			APIGroups: []string{"image.openshift.io"},
			Resources: []string{"imagestreams", "imagestreams/layers"},
			Verbs:     []string{"*"},
		})
	}

	return r, nil
}
