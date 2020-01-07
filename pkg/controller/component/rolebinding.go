package component

import (
	component "halkyon.io/api/component/v1beta1"
	"halkyon.io/operator-framework"
)

type roleBinding struct {
	framework.RoleBinding
}

func newRoleBinding(owner *component.Component) roleBinding {
	generic := framework.NewOwnedRoleBinding(owner,
		func() string { return "use-image-scc-privileged" },
		func() string { return newRole(owner).Name() },
		func() string { return ServiceAccountName(owner) })
	rb := roleBinding{RoleBinding: generic}
	return rb
}
