package controller

import (
	authorizv1 "github.com/openshift/api/authorization/v1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type role struct {
	*DependentResourceHelper
	reconciler *BaseGenericReconciler
}

func (res role) Update(toUpdate runtime.Object) (bool, error) {
	return false, nil
}

func (res role) NewInstanceWith(owner v1alpha2.Resource) DependentResource {
	return newOwnedRole(res.reconciler, owner)
}

func NewRole(reconciler *BaseGenericReconciler) role {
	return newOwnedRole(reconciler, nil)
}

func newOwnedRole(reconciler *BaseGenericReconciler, owner v1alpha2.Resource) role {
	dependent := NewDependentResource(&authorizv1.Role{}, owner)
	role := role{
		DependentResourceHelper: dependent,
		reconciler:              reconciler,
	}
	dependent.SetDelegate(role)
	return role
}

func RoleName(owner v1alpha2.Resource) string {
	return "use-scc-privileged-role"
}

func (res role) Name() string {
	return RoleName(res.Owner())
}

func (res role) Build() (runtime.Object, error) {
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

func (res role) ShouldWatch() bool {
	return false
}

func (res role) ShouldBeOwned() bool {
	return false
}

func (res role) CanBeCreatedOrUpdated() bool {
	return res.reconciler.IsTargetClusterRunningOpenShift()
}
