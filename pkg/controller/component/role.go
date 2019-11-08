package component

import (
	"halkyon.io/operator-framework"
	"halkyon.io/operator/pkg/controller"
	authorizv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type role struct {
	controller.Role
}

func newRole(owner framework.Resource) role {
	generic := controller.NewOwnedRole(owner, func() string { return "image-scc-privileged-role" })
	r := role{Role: generic}
	generic.SetDelegate(r) // we need to set the parent's delegate to the object we're creating so that its specific methods are called
	return r
}

func (res role) NewInstanceWith(owner framework.Resource) framework.DependentResource {
	return newRole(owner)
}

func (res role) Build() (runtime.Object, error) {
	ser, err := res.Role.Build()
	if err != nil {
		return nil, err
	}
	r := ser.(*authorizv1.Role)
	r.Rules = append(r.Rules, authorizv1.PolicyRule{
		APIGroups: []string{"image.openshift.io"},
		Resources: []string{"imagestreams", "imagestreams/layers"},
		Verbs:     []string{"*"},
	})

	return r, nil
}
