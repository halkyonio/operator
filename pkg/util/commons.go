package util

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
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


//GetRouteName returns an string name with the name of the router
func GetRouteName(m *v1alpha2.Component) string {
	return ""
}
