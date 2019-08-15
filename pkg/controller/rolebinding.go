package controller

import (
	"fmt"
	capability "halkyon.io/api/capability/v1beta1"
	component "halkyon.io/api/component/v1beta1"
	"halkyon.io/api/v1beta1"
	authorizv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type roleBinding struct {
	*DependentResourceHelper
}

func (res roleBinding) Update(toUpdate runtime.Object) (bool, error) {
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

func (res roleBinding) NewInstanceWith(owner v1beta1.Resource) DependentResource {
	return newOwnedRoleBinding(owner)
}

func NewRoleBinding() roleBinding {
	return newOwnedRoleBinding(nil)
}

func newOwnedRoleBinding(owner v1beta1.Resource) roleBinding {
	dependent := NewDependentResource(&authorizv1.RoleBinding{}, owner)
	rolebinding := roleBinding{
		DependentResourceHelper: dependent,
	}
	dependent.SetDelegate(rolebinding)
	return rolebinding
}

func RoleBindingName(owner v1beta1.Resource) string {
	switch owner.(type) {
	case *component.Component:
		return "use-image-scc-privileged"
	case *capability.Capability:
		return "use-scc-privileged"
	default:
		panic(fmt.Sprintf("unknown type '%s' for role owner", GetObjectName(owner)))
	}
}

func (res roleBinding) Name() string {
	return RoleBindingName(res.Owner())
}

func (res roleBinding) Build() (runtime.Object, error) {
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

	if _, ok := c.(*capability.Capability); ok {
		ser.Subjects = append(ser.Subjects, authorizv1.Subject{Kind: "ServiceAccount", Name: PostgresName(c), Namespace: namespace})
	}

	return ser, nil
}

func (res roleBinding) ShouldWatch() bool {
	return false
}
