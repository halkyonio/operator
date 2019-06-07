package component

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

//buildService returns the service resource
func (r *ReconcileComponent) buildService(res dependentResource, m *v1alpha2.Component) (runtime.Object, error) {
	ls := r.getAppLabels(m.Name)
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
	return ser, controllerutil.SetControllerReference(m, ser, r.scheme)
}
