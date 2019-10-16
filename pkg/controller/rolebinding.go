package controller

import (
	"fmt"
	"halkyon.io/operator/pkg/controller/framework"
	"halkyon.io/operator/pkg/util"
	authorizv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type RoleBinding struct {
	*framework.DependentResourceHelper
}

func (res RoleBinding) Update(toUpdate runtime.Object) (bool, error) {
	// add appropriate subject for owner
	rb := toUpdate.(*authorizv1.RoleBinding)
	owner := res.Owner()

	// check if the binding contains the current owner as subject
	namespace := owner.GetNamespace()
	name := ServiceAccountName(owner)
	found := false
	for _, subject := range rb.Subjects {
		if subject.Name == name && subject.Namespace == namespace {
			found = true
			break
		}
	}

	if !found {
		rb.Subjects = append(rb.Subjects, authorizv1.Subject{
			Kind:      "ServiceAccount",
			Namespace: namespace,
			Name:      name,
		})
	}

	return !found, nil
}

func (res RoleBinding) NewInstanceWith(owner framework.Resource) framework.DependentResource {
	return newOwnedRoleBinding(owner)
}

func NewRoleBinding() RoleBinding {
	return newOwnedRoleBinding(nil)
}

func newOwnedRoleBinding(owner framework.Resource) RoleBinding {
	dependent := framework.NewDependentResource(&authorizv1.RoleBinding{}, owner)
	rolebinding := RoleBinding{
		DependentResourceHelper: dependent,
	}
	dependent.SetDelegate(rolebinding)
	return rolebinding
}

func RoleBindingName(owner framework.Resource) string {
	switch owner.(type) {
	case *Component:
		return "use-image-scc-privileged"
	case *Capability:
		return "use-scc-privileged"
	default:
		panic(fmt.Sprintf("unknown type '%s' for role owner", util.GetObjectName(owner)))
	}
}

func (res RoleBinding) Name() string {
	return RoleBindingName(res.Owner())
}

func (res RoleBinding) Build() (runtime.Object, error) {
	c := res.Owner()
	namespace := c.GetNamespace()
	ser := &authorizv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.Name(),
			Namespace: namespace,
		},
		RoleRef: authorizv1.RoleRef{
			Kind: "Role",
			Name: RoleName(c),
		},
		Subjects: []authorizv1.Subject{
			{Kind: "ServiceAccount", Name: ServiceAccountName(c), Namespace: namespace},
		},
	}

	if _, ok := c.(*Capability); ok {
		ser.Subjects = append(ser.Subjects, authorizv1.Subject{Kind: "ServiceAccount", Name: PostgresName(c), Namespace: namespace})
	}

	return ser, nil
}

func (res RoleBinding) ShouldWatch() bool {
	return false
}
