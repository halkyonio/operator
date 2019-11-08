package component

import (
	"halkyon.io/operator-framework"
	"k8s.io/apimachinery/pkg/runtime"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type serviceAccount struct {
	base
}

func newServiceAccount(owner framework.Resource) serviceAccount {
	dependent := newBaseDependent(&corev1.ServiceAccount{}, owner)
	s := serviceAccount{base: dependent}
	dependent.SetDelegate(s)
	return s
}

//buildServiceAccount returns the service resource
func (res serviceAccount) Build() (runtime.Object, error) {
	c := res.ownerAsComponent()
	ls := getAppLabels(DeploymentName(c))
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.Name(),
			Namespace: c.Namespace,
			Labels:    ls,
		},
	}
	return sa, nil
}

func (res serviceAccount) Name() string {
	return ServiceAccountName(res.Owner())
}
