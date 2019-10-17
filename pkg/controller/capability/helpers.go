package capability

import (
	"halkyon.io/api/v1beta1"
	"strings"
)

// Convert Array of parameters to a Map
func parametersAsMap(parameters []v1beta1.NameValuePair) map[string]string {
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
