package component

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

//buildIngress returns the Ingress resource
func (r *ReconcileComponent) buildIngress(res dependentResource, c *v1alpha2.Component) (runtime.Object, error) {
	ls := getAppLabels(c.Name)
	route := &v1beta1.Ingress{
		TypeMeta: v1.TypeMeta{
			APIVersion: "networking.k8s.io/v1beta1",
			Kind:       "Ingress",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      res.name(c),
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

	// Set Component instance as the owner and controller
	return route, controllerutil.SetControllerReference(c, route, r.Scheme)
}
