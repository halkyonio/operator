package component

import (
	"context"
	"fmt"
	"github.com/halkyonio/operator/pkg/apis/component/v1alpha2"
	"github.com/halkyonio/operator/pkg/controller"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type service struct {
	base
	reconciler *ReconcileComponent // todo: remove
}

func (res service) NewInstanceWith(owner v1alpha2.Resource) controller.DependentResource {
	return newOwnedService(res.reconciler, owner)
}

func newService(reconciler *ReconcileComponent) service {
	return newOwnedService(reconciler, nil)
}

func newOwnedService(reconciler *ReconcileComponent, owner v1alpha2.Resource) service {
	dependent := newBaseDependent(&corev1.Service{}, owner)
	s := service{
		base:       dependent,
		reconciler: reconciler,
	}
	dependent.SetDelegate(s)
	return s
}

func (res service) Build() (runtime.Object, error) {
	c := res.ownerAsComponent()
	ls := getAppLabels(controller.DeploymentName(c))
	ser := &corev1.Service{
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

func (res service) Update(toUpdate runtime.Object) (bool, error) {
	c := res.ownerAsComponent()
	svc := toUpdate.(*corev1.Service)
	name := controller.DeploymentName(c)
	if svc.Spec.Selector["app"] != name {
		svc.Spec.Selector["app"] = name
		if err := res.reconciler.Client.Update(context.TODO(), svc); err != nil {
			return false, fmt.Errorf("couldn't update service '%s' selector", svc.Name)
		}
		return true, nil
	}
	return false, nil
}
