package component

import (
	"fmt"
	authorizv1 "github.com/openshift/api/authorization/v1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"github.com/snowdrop/component-operator/pkg/controller"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type rolebinding struct {
	base
	reconciler *ReconcileComponent // todo: remove
}

func (res rolebinding) NewInstanceWith(owner v1alpha2.Resource) controller.DependentResource {
	return newOwnedRoleBinding(res.reconciler, owner)
}

func newRoleBinding(reconciler *ReconcileComponent) rolebinding {
	return newOwnedRoleBinding(reconciler, nil)
}

func newOwnedRoleBinding(reconciler *ReconcileComponent, owner v1alpha2.Resource) rolebinding {
	dependent := newBaseDependent(&authorizv1.RoleBinding{}, owner)
	rolebinding := rolebinding{
		base:       dependent,
		reconciler: reconciler,
	}
	dependent.SetDelegate(rolebinding)
	return rolebinding
}

func (res rolebinding) Name() string {
	return "edit"
}

func (res rolebinding) Build() (runtime.Object, error) {
	// oc adm policy add-role-to-user edit -z build-bot
	// TODO: Fetch Edit Role and check if it exists, if not create it. This resource should be deleted if the build/tekton task is deleted (=> could be owned by tekton task maybe ?)
	c := res.ownerAsComponent()
	ser := &authorizv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "authorization.openshift.io/v1",
			Kind:       "RoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.Name(),
			Namespace: c.Namespace,
		},
		RoleRef: corev1.ObjectReference{
			Name: "edit",
		},
		Subjects: []corev1.ObjectReference{
			{Kind: "ServiceAccount", Name: serviceAccountName, Namespace: c.Namespace},
		},
		UserNames: authorizv1.OptionalNames{
			fmt.Sprintf("system:serviceaccount:%s:%s", c.Namespace, serviceAccountName),
		},
	}
	return ser, nil
}
