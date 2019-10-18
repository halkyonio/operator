package component

import (
	"halkyon.io/operator/pkg/controller"
	"halkyon.io/operator/pkg/controller/framework"
)

type roleBinding struct {
	controller.RoleBinding
}

func newRoleBinding(owner framework.Resource) roleBinding {
	rb := controller.NewOwnedRoleBinding(owner,
		func() string { return "use-image-scc-privileged" },
		func() string { return newRole(owner).Name() },
		func() string { return ServiceAccountName(owner) })
	return roleBinding{RoleBinding: rb}
}

func (res roleBinding) NewInstanceWith(owner framework.Resource) framework.DependentResource {
	return newRoleBinding(owner)
}
