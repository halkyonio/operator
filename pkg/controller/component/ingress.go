package component

import (
	"github.com/snowdrop/component-operator/pkg/controller"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type ingress struct {
	base
	reconciler *ReconcileComponent
}

func (res ingress) NewInstanceWith(owner v1.Object) controller.DependentResource {
	return newOwnedIngress(res.reconciler, owner)
}

func newIngress(reconciler *ReconcileComponent) ingress {
	return newOwnedIngress(reconciler, nil)
}

func newOwnedIngress(reconciler *ReconcileComponent, owner v1.Object) ingress {
	dependent := newBaseDependent(&v1beta1.Ingress{}, owner)
	i := ingress{base: dependent, reconciler: reconciler}
	dependent.SetDelegate(i)
	return i
}

//buildIngress returns the Ingress resource
func (res ingress) Build() (runtime.Object, error) {
	c := res.ownerAsComponent()
	ls := getAppLabels(c.Name)
	ingress := &v1beta1.Ingress{
		TypeMeta: v1.TypeMeta{
			APIVersion: "networking.k8s.io/v1beta1",
			Kind:       "Ingress",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      res.Name(),
			Namespace: c.Namespace,
			Labels:    ls,
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{Host: c.Name,
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: v1beta1.IngressBackend{
										ServiceName: c.Name,
										ServicePort: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: c.Spec.Port,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return ingress, nil
}

func (res ingress) ShouldWatch() bool {
	return !res.reconciler.isTargetClusterRunningOpenShift()
}
