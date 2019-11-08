package controller

import (
	"halkyon.io/operator-framework"
	authorizv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Role struct {
	*framework.DependentResourceHelper
	namer func() string
}

func (res Role) Update(toUpdate runtime.Object) (bool, error) {
	return false, nil
}

func NewOwnedRole(owner framework.Resource, namerFn func() string) Role {
	dependent := framework.NewDependentResource(&authorizv1.Role{}, owner)
	role := Role{DependentResourceHelper: dependent, namer: namerFn}
	dependent.SetDelegate(role)
	return role
}

func (res Role) Name() string {
	return res.namer()
}

func (res Role) Build() (runtime.Object, error) {
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

func (res Role) ShouldWatch() bool {
	return false
}
