package service

import (
	"encoding/json"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servicecatalogv1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

//buildServiceInstance returns the service resource
func (r *ReconcileService) buildServiceInstance(s *v1alpha2.Service) *servicecatalogv1.ServiceInstance {
	ls := r.GetAppLabels(s.Name)
	serviceInstanceParameters := serviceInstanceParameters(r.ParametersAsMap(s.Spec.Parameters))
	service := &servicecatalogv1.ServiceInstance{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "servicecatalog.k8s.io/v1beta1",
			Kind:       "ServiceInstance",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Name,
			Namespace: s.Namespace,
			Labels:    ls,
		},
		Spec: servicecatalogv1.ServiceInstanceSpec{
			// TODO
			PlanReference: servicecatalogv1.PlanReference{
				ClusterServiceClassExternalName: s.Spec.Class,
				ClusterServicePlanExternalName: s.Spec.Plan,
			},
			Parameters: serviceInstanceParameters,
		},
	}
	// Set Component instance as the owner and controller
	controllerutil.SetControllerReference(s, service, r.scheme)
	return service
}

// serviceInstanceParameters converts a map of variable assignments to a byte encoded json document,
// which is what the ServiceCatalog API consumes.
func serviceInstanceParameters(params map[string]string) *runtime.RawExtension {
	paramsJSON, _ := json.Marshal(params)
	return &runtime.RawExtension{Raw: paramsJSON}
}
