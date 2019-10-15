package component

import (
	"halkyon.io/operator/pkg/controller"
	"halkyon.io/operator/pkg/controller/framework"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type ingress struct {
	base
	reconciler *ComponentManager
}

func (res ingress) NewInstanceWith(owner framework.Resource) framework.DependentResource {
	return newOwnedIngress(res.reconciler, owner)
}

func newIngress(reconciler *ComponentManager) ingress {
	return newOwnedIngress(reconciler, nil)
}

func newOwnedIngress(reconciler *ComponentManager, owner framework.Resource) ingress {
	dependent := newBaseDependent(&v1beta1.Ingress{}, owner)
	i := ingress{base: dependent, reconciler: reconciler}
	dependent.SetDelegate(i)
	return i
}

func (res ingress) Build() (runtime.Object, error) {
	c := res.ownerAsComponent()
	ls := getAppLabels(controller.DeploymentName(c))
	ingress := &v1beta1.Ingress{
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
	return !res.reconciler.IsTargetClusterRunningOpenShift()
}

func (res ingress) CanBeCreatedOrUpdated() bool {
	return res.ownerAsComponent().Spec.ExposeService && res.ShouldWatch()
}
