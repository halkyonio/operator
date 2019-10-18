package component

import (
	"halkyon.io/operator/pkg/controller/framework"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type ingress struct {
	base
}

func newIngress(owner framework.Resource) ingress {
	dependent := newBaseDependent(&v1beta1.Ingress{}, owner)
	i := ingress{base: dependent}
	dependent.SetDelegate(i)
	return i
}

func (res ingress) Build() (runtime.Object, error) {
	c := res.ownerAsComponent()
	ls := getAppLabels(DeploymentName(c))
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
	return !framework.IsTargetClusterRunningOpenShift()
}

func (res ingress) CanBeCreatedOrUpdated() bool {
	return res.ownerAsComponent().Spec.ExposeService && res.ShouldWatch()
}
