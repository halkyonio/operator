package component

import (
	v1beta12 "halkyon.io/api/component/v1beta1"
	"halkyon.io/operator-framework"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type ingress struct {
	base
}

var _ framework.DependentResource = &ingress{}

func newIngress(owner *v1beta12.Component) ingress {
	config := framework.NewConfig(v1beta1.SchemeGroupVersion.WithKind("Ingress"))
	config.Watched = !framework.IsTargetClusterRunningOpenShift()
	config.Created = owner.Spec.ExposeService && config.Watched
	return ingress{base: newConfiguredBaseDependent(owner, config)}
}

func (res ingress) Build(empty bool) (runtime.Object, error) {
	ingress := &v1beta1.Ingress{}
	if !empty {
		c := res.ownerAsComponent()
		ls := getAppLabels(c)
		ingress.ObjectMeta = v1.ObjectMeta{
			Name:      res.Name(),
			Namespace: c.Namespace,
			Labels:    ls,
		}
		ingress.Spec = v1beta1.IngressSpec{
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
		}
	}

	return ingress, nil
}
