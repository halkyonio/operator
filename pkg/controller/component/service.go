package component

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type service struct {
	BaseDependentResource
}

func (p service) LabelNameFrom(object runtime.Object) string {
	return buildOrDevNamer(asComponent(object))
}

func newService(manager manager.Manager) service {
	return service{BaseDependentResource: NewDependentResource(&corev1.Service{}, manager)}
}

func (service) Name(object runtime.Object) string {
	c := asComponent(object)
	specified := c.Spec.Storage.Name
	if len(specified) > 0 {
		return specified
	}
	return "m2-data-" + c.Name
}

func (p service) Build(object runtime.Object) (runtime.Object, error) {
	m := asComponent(object)
	ls := getAppLabels(p.LabelNameFrom(m))
	ser := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.Name(m),
			Namespace: m.Namespace,
			Labels:    ls,
		},
		Spec: corev1.ServiceSpec{
			Selector: ls,
			Type:     corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: m.Spec.Port,
					},
					Port:     m.Spec.Port,
					Protocol: "TCP",
				},
			},
		},
	}
	// Set Component instance as the owner and controller
	return ser, controllerutil.SetControllerReference(m, ser, p.Scheme)
}

//buildService returns the service resource
func (r *ReconcileComponent) buildService(res dependentResource, m *v1alpha2.Component) (runtime.Object, error) {
	ls := getAppLabels(res.labelsName(m))
	ser := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.name(m),
			Namespace: m.Namespace,
			Labels:    ls,
		},
		Spec: corev1.ServiceSpec{
			Selector: ls,
			Type:     corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: m.Spec.Port,
					},
					Port:     m.Spec.Port,
					Protocol: "TCP",
				},
			},
		},
	}
	// Set Component instance as the owner and controller
	return ser, controllerutil.SetControllerReference(m, ser, r.Scheme)
}
