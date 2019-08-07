package controller

import (
	authorizv1 "k8s.io/api/rbac/v1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type k8srolebinding struct {
	*DependentResourceHelper
}

func (res k8srolebinding) Update(toUpdate runtime.Object) (bool, error) {
	return false, nil
}

func (res k8srolebinding) NewInstanceWith(owner v1alpha2.Resource) DependentResource {
	return newOwnedK8sRoleBinding(owner)
}

func NewK8sRoleBinding() k8srolebinding {
	return newOwnedK8sRoleBinding(nil)
}

func newOwnedK8sRoleBinding(owner v1alpha2.Resource) k8srolebinding {
	dependent := NewDependentResource(&authorizv1.RoleBinding{}, owner)
	rolebinding := k8srolebinding{
		DependentResourceHelper: dependent,
	}
	dependent.SetDelegate(rolebinding)
	return rolebinding
}

func (res k8srolebinding) Name() string {
	return "use-scc-privileged"
}

func (res k8srolebinding) Build() (runtime.Object, error) {
	c := res.Owner()
	namespace := c.GetNamespace()
	ser := &authorizv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.Name(),
			Namespace: namespace,
		},
		RoleRef: authorizv1.RoleRef{
			Kind:       "Role",
			Name:       RoleName(c),
		},
		Subjects: []authorizv1.Subject{
			{Kind: "ServiceAccount", Name: ServiceAccountName(c), Namespace: namespace},
		},
	}
	return ser, nil
}

func (res k8srolebinding) ShouldWatch() bool {
	return false
}
