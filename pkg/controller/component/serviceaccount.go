package component

import (
	"github.com/snowdrop/component-operator/pkg/controller"
	"k8s.io/apimachinery/pkg/runtime"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type serviceAccount struct {
	base
}

func (res serviceAccount) NewInstanceWith(owner metav1.Object) controller.DependentResource {
	return newOwnedServiceAccount(owner)
}

func newServiceAccount() serviceAccount {
	return newOwnedServiceAccount(nil)
}

func newOwnedServiceAccount(owner metav1.Object) serviceAccount {
	dependent := newBaseDependent(&corev1.ServiceAccount{}, owner)
	s := serviceAccount{base: dependent}
	dependent.SetDelegate(s)
	return s
}

//buildServiceAccount returns the service resource
func (res serviceAccount) Build() (runtime.Object, error) {
	c := res.ownerAsComponent()
	ls := getAppLabels(c.Name)
	sa := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.Name(),
			Namespace: c.Namespace,
			Labels:    ls,
		},
	}
	return sa, nil
}

func (res serviceAccount) Name() string {
	return serviceAccountName
}
