package util

import (
	"fmt"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"os"
)

const (
	// WatchNamespaceEnvVar is the constant for env variable WATCH_NAMESPACE
	// which is the namespace that the pod is currently running in.
	WatchNamespaceEnvVar = "WATCH_NAMESPACE"
)

func IsOpenshift(kubeconfig *rest.Config) (bool,error) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(kubeconfig)
	if err != nil {
		return false, err
	}
	apiList, err := discoveryClient.ServerGroups()
	if err != nil {
		return false, err
	}
	apiGroups := apiList.Groups
	for i := 0; i < len(apiGroups); i++ {
		if (apiGroups[i].Name == "route.openshift.io") {
			return true, nil
		}
	}
	return false, nil
}

// GetWatchNamespace returns the namespace the operator should be watching for changes
func GetWatchNamespace() (string, error) {
	ns, found := os.LookupEnv(WatchNamespaceEnvVar)
	if !found {
		return "", fmt.Errorf("%s must be set", WatchNamespaceEnvVar)
	}
	return ns, nil
}

//GetRouteName returns an string name with the name of the router
func GetRouteName(m *v1alpha2.Component) string {
	return ""
}
