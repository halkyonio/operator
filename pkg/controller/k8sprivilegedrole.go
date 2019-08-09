package controller

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	authorizv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type k8sprivilegedrole struct {
	*DependentResourceHelper
}

func (res k8sprivilegedrole) Update(toUpdate runtime.Object) (bool, error) {
	return false, nil
}

func (res k8sprivilegedrole) NewInstanceWith(owner v1alpha2.Resource) DependentResource {
	return newOwnedK8sPrivilegedRole(owner)
}

func NewK8sPrivilegedRole() k8sprivilegedrole {
	return newOwnedK8sPrivilegedRole(nil)
}

func newOwnedK8sPrivilegedRole(owner v1alpha2.Resource) k8sprivilegedrole {
	dependent := NewDependentResource(&authorizv1.Role{}, owner)
	role := k8sprivilegedrole{DependentResourceHelper: dependent}
	dependent.SetDelegate(role)
	return role
}

func PrivilegedRoleName(owner v1alpha2.Resource) string {
	return "scc-privileged-role"
}

func (res k8sprivilegedrole) Name() string {
	return PrivilegedRoleName(res.Owner())
}

func (res k8sprivilegedrole) Build() (runtime.Object, error) {
	c := res.Owner()
	ser := &authorizv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.Name(),
			Namespace: c.GetNamespace(),
		},
		Rules: []authorizv1.PolicyRule{
			{
				APIGroups:     []string{"security.openshift.io"},
				Resources:     []string{"securitycontextconstraints"},
				ResourceNames: []string{"privileged"},
				Verbs:         []string{"use"},
			},
		},
	}
	return ser, nil
}

func (res k8sprivilegedrole) ShouldWatch() bool {
	return false
}
