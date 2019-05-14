package service

import (
	"encoding/json"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
)

// BuildParameters converts a map of variable assignments to a byte encoded json document,
// which is what the ServiceCatalog API consumes.
func (r *ReconcileService) BuildParameters(params interface{}) *runtime.RawExtension {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		// This should never be hit because marshalling a map[string]string is pretty safe
		// I'd rather throw a panic then force handling of an error that I don't think is possible.
		r.reqLogger.Error(err,"unable to marshal the request parameters")
	}
	return &runtime.RawExtension{Raw: paramsJSON}
}

// Convert Array of parameters to a Map
func (r *ReconcileService) ParametersAsMap(parameters []v1alpha2.Parameter) map[string]string {
	result := make(map[string]string)
	for _, parameter := range parameters {
		result[parameter.Name] = parameter.Value
	}
	return result
}
