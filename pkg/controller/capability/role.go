package capability

import (
	"halkyon.io/operator/pkg/controller"
	"halkyon.io/operator/pkg/controller/framework"
)

type role struct {
	controller.Role
}

func newRole(owner framework.Resource) role {
	r := controller.NewOwnedRole(owner, func() string { return "scc-privileged-role" })
	return role{Role: r}
}
