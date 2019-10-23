package component

import (
	"fmt"
	"halkyon.io/api/component/v1beta1"
	halkyon "halkyon.io/api/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	supervisorContainerName = "copy-supervisord"
	supervisorImageId       = "supervisord"
	latestVersionTag        = "latest"
)

// todo: extract into configuration file
var registry = RuntimeRegistry{
	runtimes: map[string]runtimeVersions{
		"spring-boot":     newRuntime("quay.io/halkyonio/hal-maven-jdk", "*"),
		"vert.x":          newRuntime("quay.io/halkyonio/hal-maven-jdk", "*-all"),
		"quarkus":         newRuntime("quay.io/halkyonio/hal-maven-jdk", "*-runner"),
		"thorntail":       newRuntime("quay.io/halkyonio/hal-maven-jdk", "*"),
		"openjdk8":        newRuntime("registry.access.redhat.com/redhat-openjdk-18/openjdk18-openshift", ""),
		"node.js":         newRuntime("nodeshift/centos7-s2i-nodejs", ""),
		supervisorImageId: newRuntime("quay.io/halkyonio/supervisord", ""),
	},
}

type Runtime struct {
	RegistryRef string
	defaultEnv  map[string]string
}

type RuntimeRegistry struct {
	runtimes map[string]runtimeVersions
}

type runtimeVersions struct {
	defaultVersion Runtime
}

func newRuntime(ref, jarPattern string) runtimeVersions {
	return runtimeVersions{defaultVersion: Runtime{RegistryRef: ref, defaultEnv: map[string]string{"JARPATTERN": jarPattern}}}
}

func (rv runtimeVersions) getRuntimeVersionFor(version string) Runtime {
	// todo: compute image version based on runtime version
	defaultVersion := rv.defaultVersion
	ref := fmt.Sprintf("%s:%s", defaultVersion.RegistryRef, latestVersionTag)
	runtime := Runtime{RegistryRef: ref, defaultEnv: defaultVersion.defaultEnv}
	return runtime
}

func getImageInfo(component v1beta1.ComponentSpec) (Runtime, error) {
	versions, ok := registry.runtimes[component.Runtime]
	if !ok {
		return Runtime{}, fmt.Errorf("unknown image identifier: %s", component.Runtime)
	}

	return versions.getRuntimeVersionFor(component.Version), nil
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
