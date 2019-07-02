package component

import (
	"k8s.io/apimachinery/pkg/runtime"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type serviceAccount struct {
	base
}

func newServiceAccount() serviceAccount {
	return serviceAccount{base: newBaseDependent(&corev1.ServiceAccount{})}
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
