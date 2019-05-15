package capability

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servicecatalogv1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

//buildServiceInstance returns the service resource
func (r *ReconcileCapability) buildServiceInstance(s *v1alpha2.Capability) (*servicecatalogv1.ServiceInstance, error) {
	ls := r.GetAppLabels(s.Name)
	serviceInstanceParameters, err := convertParametersToMap(r.ParametersAsMap(s.Spec.Parameters))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create the service instance parameters")
	}

	service := &servicecatalogv1.ServiceInstance{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "servicecatalog.k8s.io/v1beta1",
			Kind:       "ServiceInstance",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Spec.Name,
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
	return service, nil
}

// serviceInstanceParameters converts a map of variable assignments to a byte encoded json document,
// which is what the ServiceCatalog API consumes.
func convertParametersToMap(params map[string]string) (*runtime.RawExtension, error) {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	return &runtime.RawExtension{Raw: paramsJSON}, nil
}
