package controller

import (
	authorizv1 "k8s.io/api/rbac/v1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type k8sprivilegedrolebinding struct {
	*DependentResourceHelper
}

func (res k8sprivilegedrolebinding) Update(toUpdate runtime.Object) (bool, error) {
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

func (res k8sprivilegedrolebinding) NewInstanceWith(owner v1alpha2.Resource) DependentResource {
	return newOwnedK8sPrivilegedRoleBinding(owner)
}

func NewK8sPrivilegedRoleBinding() k8sprivilegedrolebinding {
	return newOwnedK8sPrivilegedRoleBinding(nil)
}

func newOwnedK8sPrivilegedRoleBinding(owner v1alpha2.Resource) k8sprivilegedrolebinding {
	dependent := NewDependentResource(&authorizv1.RoleBinding{}, owner)
	rolebinding := k8sprivilegedrolebinding{
		DependentResourceHelper: dependent,
	}
	dependent.SetDelegate(rolebinding)
	return rolebinding
}

func (res k8sprivilegedrolebinding) Name() string {
	return "use-scc-privileged"
}

func (res k8sprivilegedrolebinding) Build() (runtime.Object, error) {
	c := res.Owner()
	namespace := c.GetNamespace()
	ser := &authorizv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.Name(),
			Namespace: namespace,
		},
		RoleRef: authorizv1.RoleRef{
			Kind: "Role",
			Name: ImageAndPrivilegedRoleName(c),
		},
		Subjects: []authorizv1.Subject{
			{Kind: "ServiceAccount", Name: ServiceAccountName(c), Namespace: namespace},
			{Kind: "ServiceAccount", Name: PostgresName(c), Namespace: namespace},
		},
	}
	return ser, nil
}

func (res k8sprivilegedrolebinding) ShouldWatch() bool {
	return false
}
