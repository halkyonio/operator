package component

import (
	"fmt"
	"halkyon.io/api/component/v1beta1"
	"halkyon.io/api/runtime/clientset/versioned"
	v1beta12 "halkyon.io/api/runtime/clientset/versioned/typed/runtime/v1beta1"
	halkyon "halkyon.io/api/v1beta1"
	framework "halkyon.io/operator-framework"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	supervisorContainerName = "copy-supervisord"
	supervisorImageId       = "supervisord"
)

var runtimesClient v1beta12.RuntimeInterface

type Runtime struct {
	RegistryRef string
	defaultEnv  map[string]string
}

func getImageInfo(component v1beta1.ComponentSpec) (Runtime, error) {
	if component.Runtime == supervisorImageId {
		return Runtime{RegistryRef: "quay.io/halkyonio/supervisord"}, nil
	}

	if runtimesClient == nil {
		runtimesClient = versioned.NewForConfigOrDie(framework.Helper.Config).HalkyonV1beta1().Runtimes()
	}
	list, err := runtimesClient.List(v1.ListOptions{})
	if err != nil {
		return Runtime{}, fmt.Errorf("couldn't retrieve available runtimes: %e", err)
	}

	runtimeFound := false
	knownVersions := ""
	knownRuntimes := ""
	for _, item := range list.Items {
		if item.Spec.Name == component.Runtime {
			runtimeFound = true
			version := item.Spec.Version
			if version == component.Version {
				runtime := Runtime{RegistryRef: item.Spec.Image}
				if len(item.Spec.ExecutablePattern) > 0 {
					runtime.defaultEnv = map[string]string{"JARPATTERN": item.Spec.ExecutablePattern}
				}
				return runtime, nil
			}
			knownVersions += version + ", "
		}
		knownRuntimes += item.Name + ", "
	}

	if runtimeFound {
		return Runtime{}, fmt.Errorf("couldn't find '%s' version for '%s' runtime, known versions: %s", component.Version, component.Runtime, knownVersions)
	}
	return Runtime{}, fmt.Errorf("couldn't find '%s' runtime, known runtimes: %s", component.Runtime, knownRuntimes)
}

func getSupervisor() *v1beta1.Component {
	return &v1beta1.Component{
		ObjectMeta: v1.ObjectMeta{
			Name: supervisorContainerName,
		},
		Spec: v1beta1.ComponentSpec{
			Runtime: supervisorImageId,
			Envs: []halkyon.NameValuePair{
				{
					Name: "CMDS",
					Value: "build:/usr/local/bin/build:false;" +
						"run:/usr/local/bin/run:true",
				},
			},
		},
	}
}
