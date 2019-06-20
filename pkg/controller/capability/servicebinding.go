package capability

import (
	servicecatalogv1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

//buildServiceBinding returns the service binding resource
func (r *ReconcileCapability) buildServiceBinding(s *v1alpha2.Capability) *servicecatalogv1.ServiceBinding {
	ls := r.GetAppLabels(s.Name)
	service := &servicecatalogv1.ServiceBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "servicecatalog.k8s.io/v1beta1",
			Kind:       "ServiceBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Name,
			Namespace: s.Namespace,
			Labels:    ls,
		},
		Spec: servicecatalogv1.ServiceBindingSpec{
			SecretName:         s.Spec.SecretName,
			ServiceInstanceRef: servicecatalogv1.LocalObjectReference{Name: s.Name},
		},
	}
	// Set Component instance as the owner and controller
	controllerutil.SetControllerReference(s, service, r.Scheme)
	return service
}
