package util

import (
	"fmt"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

func IsOpenshift(kubeconfig *rest.Config) (bool, error) {
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
		if apiGroups[i].Name == "route.openshift.io" {
			return true, nil
		}
	}
	return false, nil
}

func GetImageReference(imageName string, version ...string) string {
	runtimeVersion := "latest"
	if len(version) == 1 && len(version[0]) > 0 {
		runtimeVersion = version[0]
	}
	return fmt.Sprintf("%s:%s", imageName, runtimeVersion)
}
