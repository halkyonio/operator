package component

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

//buildServiceAccount returns the service resource
func (r *ReconcileComponent) buildServiceAccount(res dependentResource, m *v1alpha2.Component) (runtime.Object, error) {
	ls := r.getAppLabels(m.Name)
	sa := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.name(m),
			Namespace: m.Namespace,
			Labels:    ls,
		},
	}
	// Set Component instance as the owner and controller
	return sa, controllerutil.SetControllerReference(m, sa, r.Scheme)
}
