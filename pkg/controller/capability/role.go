package capability

import (
	"halkyon.io/operator/pkg/controller"
	"halkyon.io/operator/pkg/controller/framework"
)

type role struct {
	controller.Role
}

func newRole(owner framework.Resource) role {
	generic := controller.NewOwnedRole(owner, func() string { return "scc-privileged-role" })
	r := role{Role: generic}
	generic.SetDelegate(r)
	return r
}
