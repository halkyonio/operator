package component

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type service struct {
	base
	reconciler *ReconcileComponent // todo: remove
}

func newService(reconciler *ReconcileComponent) service {
	return service{
		base:       newBaseDependent(&corev1.Service{}),
		reconciler: reconciler,
	}
}

func (res service) Build() (runtime.Object, error) {
	c := res.ownerAsComponent()
	ls := getAppLabels(buildOrDevNamer(c))
	ser := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.Name(),
			Namespace: c.Namespace,
			Labels:    ls,
		},
		Spec: corev1.ServiceSpec{
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
		},
	}
	return ser, nil
}

func (res service) Update(toUpdate metav1.Object) (bool, error) {
	c := res.ownerAsComponent()
	svc := toUpdate.(*corev1.Service)
	name := buildOrDevNamer(c)
	if svc.Spec.Selector["app"] != name {
		svc.Spec.Selector["app"] = name
		if err := res.reconciler.Client.Update(context.TODO(), svc); err != nil {
			return false, fmt.Errorf("couldn't update service '%s' selector", svc.Name)
		}
		return true, nil
	}
	return false, nil
}
