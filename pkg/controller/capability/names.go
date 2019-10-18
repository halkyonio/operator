package capability

import (
	"halkyon.io/operator/pkg/controller/framework"
)

func PostgresName(owner framework.Resource) string {
	return framework.DefaultDependentResourceNameFor(owner)
}

func ServiceAccountName(owner framework.Resource) string {
	return PostgresName(owner) // todo: fix me
}
