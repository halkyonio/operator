package component

import (
	authorizv1 "github.com/openshift/api/authorization/v1"
	"github.com/snowdrop/component-operator/pkg/controller"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type rolebinding struct {
	base
	reconciler *ReconcileComponent // todo: remove
}

func (res rolebinding) NewInstanceWith(owner metav1.Object) controller.DependentResource {
	return newOwnedRoleBinding(res.reconciler, owner)
}

func newRoleBinding(reconciler *ReconcileComponent) rolebinding {
	return newOwnedRoleBinding(reconciler, nil)
}

func newOwnedRoleBinding(reconciler *ReconcileComponent, owner metav1.Object) rolebinding {
	dependent := newBaseDependent(&corev1.Service{}, owner)
	rolebinding := rolebinding{
		base:       dependent,
		reconciler: reconciler,
	}
	dependent.SetDelegate(rolebinding)
	return rolebinding
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
			Name: "edit",
			Namespace: c.Namespace,
		},
		RoleRef: corev1.ObjectReference{
			Name: "edit",
		},
		Subjects: []corev1.ObjectReference{
			{Kind: "ServiceAccount", Name: "Tekton ServiceAccoiunt -> build-bot", Namespace: c.Namespace},
		},
		UserNames: authorizv1.OptionalNames{
			"system:serviceaccount:<NAMESPACE>:<TEKTON SA -> build-bot>",
		},
	}
	return ser, nil
}



// oc get rolebinding/edit -n gytis-test -o yaml
//apiVersion: authorization.openshift.io/v1
//groupNames: null
//kind: RoleBinding
//metadata:
//  creationTimestamp: 2019-07-15T10:45:44Z
//  name: edit
//  namespace: gytis-test
//  resourceVersion: "45987567"
//  selfLink: /apis/authorization.openshift.io/v1/namespaces/gytis-test/rolebindings/edit
//  uid: b2c34b19-a6ed-11e9-bbd9-107b44b03540
//roleRef:
//  name: edit
//subjects:
//- kind: ServiceAccount
//  name: build-bot
//  namespace: gytis-test
//userNames:
//- system:serviceaccount:gytis-test:build-bot