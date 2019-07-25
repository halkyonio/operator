package capability

import (
	"encoding/json"
	"github.com/snowdrop/component-api/pkg/apis/component/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
	"strings"
)

// BuildParameters converts a map of variable assignments to a byte encoded json document,
// which is what the ServiceCatalog API consumes.
func (r *ReconcileCapability) BuildParameters(params interface{}) *runtime.RawExtension {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		// This should never be hit because marshalling a map[string]string is pretty safe
		// I'd rather throw a panic then force handling of an error that I don't think is possible.
		r.ReqLogger.Error(err, "unable to marshal the request parameters")
	}
	return &runtime.RawExtension{Raw: paramsJSON}
}

// Convert Array of parameters to a Map
func parametersAsMap(parameters []v1alpha2.Parameter) map[string]string {
	result := make(map[string]string)
	for _, parameter := range parameters {
		result[parameter.Name] = parameter.Value
	}
	return result
}

func SetDefaultSecretNameIfEmpty(capabilityName, paramSecretName string) string {
	if paramSecretName == "" {
		return strings.ToLower(capabilityName) + "-config"
	} else {
		return paramSecretName
	}
}

func SetDefaultDatabaseName(paramDatabaseName string) string {
	if paramDatabaseName == "" {
		return "sample-db"
	} else {
		return paramDatabaseName
	}
}

func SetDefaultDatabaseHost(capabilityHost, paramHost string) string {
	if paramHost == "" {
		return capabilityHost
	} else {
		return paramHost
	}
}

func SetDefaultDatabasePort(paramPort string) string {
	// TODO. Assign port according to the DB type using Enum
	if paramPort == "" {
		return "5432"
	} else {
		return paramPort
	}
}

//getAppLabels returns an string map with the labels which wil be associated to the kubernetes/ocp resource which will be created and managed by this operator
func getAppLabels(name string) map[string]string {
	return map[string]string{
		"app": name,
	}
}

func (r *ReconcileCapability) ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func (r *ReconcileCapability) RemoveString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}
