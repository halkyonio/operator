package main

import (
	"flag"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"strings"

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
	dc, err := dynamic.NewForConfig(cfg)
	if err != nil {
		panic(err.Error())
	}

	// Get Server Resources
	serverResources, err := c.Discovery().ServerResources()
	if err != nil {
		panic(err.Error())
	}

	gvksMap, err := discovery.GroupVersionResources(serverResources)
	gvks := filterGvks(gvksMap)

	for _, gvk := range gvks {
		list, err := dc.Resource(gvk).Namespace("default").List(metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%v=%v", "docker-registry", "default"),
		})
		if err == nil && len(list.Items) > 0 {
			fmt.Printf("%v\n", list)
		}

	}

}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

var allowedGroups = []string{"build.openshift.io", "image.openshift.io", "route.openshift.io", "component.k8s.io", "apps", "apps.openshift.io"}
var allowedCoreResources = []string{"pod", "replicationcontrollers", "services"}

func filterGvks(input map[schema.GroupVersionResource]struct{}) []schema.GroupVersionResource {
	result := []schema.GroupVersionResource{}

	for gvk := range input {
		if strings.Contains(gvk.Resource, "/") {
			continue
		}
		add := false
		if gvk.Group == "" {
			for _, res := range allowedCoreResources {
				if res == gvk.Resource {
					add = true
				}
			}
		} else {
			for _, grp := range allowedGroups {
				if grp == gvk.Group {
					add = true
				}
			}
		}
		if add {
			result = append(result, gvk)
		}
	}

	return result
}
