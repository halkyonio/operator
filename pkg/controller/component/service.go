package component

import (
	v1beta12 "halkyon.io/api/component/v1beta1"
	"halkyon.io/operator-framework"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type service struct {
	base
}

var _ framework.DependentResource = &service{}
var serviceGVK = corev1.SchemeGroupVersion.WithKind("Service")

func newService(owner *v1beta12.Component) service {
	config := framework.NewConfig(serviceGVK)
	config.CheckedForReadiness = true
	config.Updated = true
	return service{base: newConfiguredBaseDependent(owner, config)}
}

func (res service) Build(empty bool) (runtime.Object, error) {
	ser := &corev1.Service{}
	if !empty {
		c := res.ownerAsComponent()
		ls := getAppLabels(c)
		ser.ObjectMeta = metav1.ObjectMeta{
			Name:      res.Name(),
			Namespace: c.Namespace,
			Labels:    ls,
		}
		ser.Spec = corev1.ServiceSpec{
			Selector: ls,
			Type:     corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: c.Spec.Port,
					},
					Port:     c.Spec.Port,
					Protocol: "TCP",
				},
			},
		}
	}
	return ser, nil
}

func (res service) Update(toUpdate runtime.Object) (bool, runtime.Object, error) {
	c := res.ownerAsComponent()
	svc := toUpdate.(*corev1.Service)
	labels := getAppLabels(c)
	if svc.Spec.Selector["app"] != labels["app"] {
		for key, value := range labels {
			svc.Spec.Selector[key] = value
		}
		return true, toUpdate, nil
	}
	return false, toUpdate, nil
}
