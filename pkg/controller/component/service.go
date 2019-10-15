package component

import (
	"halkyon.io/operator/pkg/controller"
	"halkyon.io/operator/pkg/controller/framework"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type service struct {
	base
	reconciler *ComponentManager // todo: remove
}

func (res service) NewInstanceWith(owner framework.Resource) framework.DependentResource {
	return newOwnedService(res.reconciler, owner)
}

func newService(reconciler *ComponentManager) service {
	return newOwnedService(reconciler, nil)
}

func newOwnedService(reconciler *ComponentManager, owner framework.Resource) service {
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
		labels := getAppLabels(name)
		for key, value := range labels {
			svc.Spec.Selector[key] = value
		}
		return true, nil
	}
	return false, nil
}
