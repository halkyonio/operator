package capability

import (
	"halkyon.io/operator/pkg/controller"
	"halkyon.io/operator/pkg/controller/framework"
	authorizv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type roleBinding struct {
	controller.RoleBinding
}

func newRoleBinding(owner framework.Resource) roleBinding {
	rb := controller.NewOwnedRoleBinding(owner,
		func() string { return "use-scc-privileged" },
		func() string { return newRole(owner).Name() },
		func() string { return newOwnedPostgres(owner).Name() })
	return roleBinding{RoleBinding: rb}
}

func (res roleBinding) NewInstanceWith(owner framework.Resource) framework.DependentResource {
	return newRoleBinding(owner)
}

func (res roleBinding) Build() (runtime.Object, error) {
	build, err := res.RoleBinding.Build()
	if err != nil {
		return nil, err
	}

	ser := build.(*authorizv1.RoleBinding)
	owner := res.Owner()
	ser.Subjects = append(ser.Subjects, authorizv1.Subject{Kind: "ServiceAccount", Name: PostgresName(owner), Namespace: owner.GetNamespace()})

	return ser, nil
}
