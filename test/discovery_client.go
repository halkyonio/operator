package main

import (
	"flag"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/dynamic"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"path/filepath"
	"sort"
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

	// Print Server Group Version
	serverGroupVersions := groupVersions(serverResources)
	sort.Strings(serverGroupVersions)
	for _, serverGroupVersion := range serverGroupVersions {
		fmt.Println("Item : ", serverGroupVersion)

		// TODO
		dc, err := dynamic.NewForConfig(cfg)
		if err != nil {
			panic(err.Error())
		}
		gvk := &schema.GroupVersionResource{
			Resource: "",
			Group: "",
			Version: "",
		}
		opts := metav1.ListOptions{}
		_, err = dc.Resource(*gvk).List(opts)
		if err != nil {
			panic(err.Error())
		}
		// fmt.Println(t)
	}

	// Print API resource
	for _, apiResourceList := range serverResources {
		apiResources := apiResourceList.APIResources
		for _, gvk := range apiResources {
			fmt.Printf("Group, Version, Kind : %s, %s, %s\n",gvk.Group, gvk.Version, gvk.Kind)
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

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

