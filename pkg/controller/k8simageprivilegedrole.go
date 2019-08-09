package controller

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	authorizv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type k8simageprivilegedrole struct {
	*DependentResourceHelper
}

func (res k8simageprivilegedrole) Update(toUpdate runtime.Object) (bool, error) {
	return false, nil
}

func (res k8simageprivilegedrole) NewInstanceWith(owner v1alpha2.Resource) DependentResource {
	return newOwnedK8sImageAndPrivilegedRole(owner)
}

func NewK8sImageAndPrivilegedRole() k8simageprivilegedrole {
	return newOwnedK8sImageAndPrivilegedRole(nil)
}

func newOwnedK8sImageAndPrivilegedRole(owner v1alpha2.Resource) k8simageprivilegedrole {
	dependent := NewDependentResource(&authorizv1.Role{}, owner)
	role := k8simageprivilegedrole{DependentResourceHelper: dependent}
	dependent.SetDelegate(role)
	return role
}

func ImageAndPrivilegedRoleName(owner v1alpha2.Resource) string {
	return "image-scc-privileged-role"
}

func (res k8simageprivilegedrole) Name() string {
	return ImageAndPrivilegedRoleName(res.Owner())
}

func (res k8simageprivilegedrole) Build() (runtime.Object, error) {
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
			{
				APIGroups:     []string{"image.openshift.io"},
				Resources:     []string{"imagestreams","imagestreams/layers"},
				Verbs:         []string{"*"},
			},
		},
	}
	return ser, nil
}

func (res k8simageprivilegedrole) ShouldWatch() bool {
	return false
}
