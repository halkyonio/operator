package main

import (
	"flag"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"strings"

	"k8s.io/client-go/kubernetes"
	//"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/clientcmd"
	//"k8s.io/client-go/dynamic"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"path/filepath"
)

var (
	scheme = runtime.NewScheme()
)

func main() {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	cfg, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	c, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		panic(err.Error())
	}
	// Get Server Resources
	serverResources, err := c.Discovery().ServerResources()
	if err != nil {
		panic(err.Error())
	}

	for _, serverResource := range serverResources {
		groupVersion := serverResource.GroupVersion
		s := strings.Split(groupVersion, "/")
		parentGroup := ""
		parentVersion := ""
		if len(s) == 1 {
			parentVersion = s[0]
		} else if len(s) == 2 {
			parentGroup = s[0]
			parentVersion = s[1]
		}
		kindsOfServerResource := []string{}
		for _, apiResource := range serverResource.APIResources {
			effectiveVersion := apiResource.Version
			if effectiveVersion == "" {
				effectiveVersion = parentVersion
			}
			effectiveGroup := apiResource.Group
			if effectiveGroup == "" {
				effectiveGroup = parentGroup
			}

			kind := apiResource.Kind
			found := false
			// make sure we don't print duplicates
			for _, alreadySeenKind := range kindsOfServerResource {
				if alreadySeenKind == kind {
					found = true
					break
				}
			}

			if !found {
				kindsOfServerResource = append(kindsOfServerResource, kind)
				fmt.Printf("Group, Version, Kind : %s, %s, %s\n", effectiveGroup, effectiveVersion, apiResource.Kind)
			}
		}
	}

}

func groupVersions(resources []*metav1.APIResourceList) []string {
	result := []string{}
	for _, resourceList := range resources {
		result = append(result, resourceList.GroupVersion)
	}
	return result
}

func gvkList(apiList []*metav1.APIResourceList) [][]metav1.APIResource {
	result := [][]metav1.APIResource{}
	for _, resourceList := range apiList {
		result = append(result, resourceList.APIResources)
	}
	return result
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
