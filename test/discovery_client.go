package main

import (
	"flag"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
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
	dc, err := dynamic.NewForConfig(cfg)
	if err != nil {
		panic(err.Error())
	}

	// Get Server Resources
	serverResources, err := c.Discovery().ServerResources()
	if err != nil {
		panic(err.Error())
	}

	gvks, err := discovery.GroupVersionResources(serverResources)

	for gvk := range gvks {
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
