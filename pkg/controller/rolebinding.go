package controller

import (
	authorizv1 "github.com/openshift/api/authorization/v1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type rolebinding struct {
	*DependentResourceHelper
	reconciler *BaseGenericReconciler
}

func (res rolebinding) Update(toUpdate runtime.Object) (bool, error) {
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
		rb.Subjects = append(rb.Subjects, corev1.ObjectReference{
			Kind:      "ServiceAccount",
			Namespace: namespace,
			Name:      name,
		})
	}

	return !found, nil
}

func (res rolebinding) NewInstanceWith(owner v1alpha2.Resource) DependentResource {
	return newOwnedRoleBinding(res.reconciler, owner)
}

func NewRoleBinding(reconciler *BaseGenericReconciler) rolebinding {
	return newOwnedRoleBinding(reconciler, nil)
}

func newOwnedRoleBinding(reconciler *BaseGenericReconciler, owner v1alpha2.Resource) rolebinding {
	dependent := NewDependentResource(&authorizv1.RoleBinding{}, owner)
	rolebinding := rolebinding{
		DependentResourceHelper: dependent,
		reconciler:              reconciler,
	}
	dependent.SetDelegate(rolebinding)
	return rolebinding
}

func (res rolebinding) Name() string {
	return "use-scc-privileged"
}

func (res rolebinding) Build() (runtime.Object, error) {
	c := res.Owner()
	namespace := c.GetNamespace()
	ser := &authorizv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.Name(),
			Namespace: namespace,
		},
		RoleRef: corev1.ObjectReference{
			Kind:       "Role",
			Name:       RoleName(c),
			APIVersion: "rbac.authorization.k8s.io/v1beta1",
			Namespace:  namespace,
		},
		// todo: the second service account should be added on demand in Update but for some reason the server doesn't accept it
		//  despite returning 200OK on the operation… :/
		Subjects: []corev1.ObjectReference{
			{Kind: "ServiceAccount", Name: ServiceAccountName(c), Namespace: namespace},
			{Kind: "ServiceAccount", Name: PostgresName(c), Namespace: namespace},
		},
	}
	return ser, nil
}

func (res rolebinding) ShouldWatch() bool {
	return false
}

func (res rolebinding) ShouldBeOwned() bool {
	return false
}

func (res rolebinding) CanBeCreatedOrUpdated() bool {
	return res.reconciler.IsTargetClusterRunningOpenShift()
}
