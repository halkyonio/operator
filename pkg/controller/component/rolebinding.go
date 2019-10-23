package component

import (
	"halkyon.io/operator/pkg/controller"
	"halkyon.io/operator/pkg/controller/framework"
)

type roleBinding struct {
	controller.RoleBinding
}

func newRoleBinding(owner framework.Resource) roleBinding {
	generic := controller.NewOwnedRoleBinding(owner,
		func() string { return "use-image-scc-privileged" },
		func() string { return newRole(owner).Name() },
		func() string { return ServiceAccountName(owner) })
	rb := roleBinding{RoleBinding: generic}
	generic.SetDelegate(rb)
	return rb
}

func (res roleBinding) NewInstanceWith(owner framework.Resource) framework.DependentResource {
	return newRoleBinding(owner)
}
