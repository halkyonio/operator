package component

import (
	"halkyon.io/api/v1beta1"
	framework "halkyon.io/operator-framework"
	"k8s.io/apimachinery/pkg/runtime"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type serviceAccount struct {
	base
}

var _ framework.DependentResource = &serviceAccount{}

func newServiceAccount(owner v1beta1.HalkyonResource) serviceAccount {
	return serviceAccount{base: newBaseDependent(&corev1.ServiceAccount{}, owner)}
}

//buildServiceAccount returns the service resource
func (res serviceAccount) Build(empty bool) (runtime.Object, error) {
	sa := &corev1.ServiceAccount{}
	if !empty {
		c := res.ownerAsComponent()
		ls := getAppLabels(DeploymentName(c))
		sa.ObjectMeta = metav1.ObjectMeta{
			Name:      res.Name(),
			Namespace: c.Namespace,
			Labels:    ls,
		}
	}
	return sa, nil
}

func (res serviceAccount) Name() string {
	return ServiceAccountName(res.Owner())
}
