/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package component

import (
	"context"
	component "halkyon.io/api/component/v1beta1"
	"halkyon.io/api/v1beta1"
	controller2 "halkyon.io/operator/pkg/controller"
	"halkyon.io/operator/pkg/controller/framework"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewComponentManager() *ComponentManager {
	// todo: make this configurable
	images := make(map[string]imageInfo, 7)
	defaultEnvVar := make(map[string]string, 7)
	// defaultEnvVar["JAVA_APP_JAR"] = "app.jar"
	images["spring-boot"] = imageInfo{
		registryRef: "quay.io/halkyonio/hal-maven-jdk",
		defaultEnv:  defaultEnvVar,
	}
	images["vert.x"] = imageInfo{
		registryRef: "quay.io/halkyonio/hal-maven-jdk",
		defaultEnv:  defaultEnvVar,
	}
	images["quarkus"] = imageInfo{
		registryRef: "quay.io/halkyonio/hal-maven-jdk",
		defaultEnv:  defaultEnvVar,
	}
	images["thorntail"] = imageInfo{
		registryRef: "quay.io/halkyonio/hal-maven-jdk",
		defaultEnv:  defaultEnvVar,
	}
	// References images
	images["openjdk8"] = imageInfo{registryRef: "registry.access.redhat.com/redhat-openjdk-18/openjdk18-openshift"}
	images["node.js"] = imageInfo{registryRef: "nodeshift/centos7-s2i-nodejs"}
	images[supervisorImageId] = imageInfo{registryRef: "quay.io/halkyonio/supervisord"}

	supervisor := component.Component{
		ObjectMeta: v1.ObjectMeta{
			Name: supervisorContainerName,
		},
		Spec: component.ComponentSpec{
			Runtime: supervisorImageId,
			Version: latestVersionTag,
			Envs: []v1beta1.NameValuePair{
				{
					Name: "CMDS",
					Value: "build:/usr/local/bin/build:false;" +
						"run:/usr/local/bin/run:true",
				},
			},
		},
	}

	r := &ComponentManager{
		runtimeImages: images,
		supervisor:    &supervisor,
	}
	return r
}

type imageInfo struct {
	registryRef string
	defaultEnv  map[string]string
}

type ComponentManager struct {
	*framework.K8SHelper
	runtimeImages  map[string]imageInfo
	supervisor     *component.Component
	dependentTypes []framework.DependentResource
}

func (r *ComponentManager) SetHelper(helper *framework.K8SHelper) {
	r.K8SHelper = helper
}

func (r *ComponentManager) Helper() *framework.K8SHelper {
	return r.K8SHelper
}

func (r *ComponentManager) GetDependentResourcesTypes() []framework.DependentResource {
	if len(r.dependentTypes) == 0 {
		r.dependentTypes = []framework.DependentResource{
			newPvc(),
			newDeployment(r),
			newService(r),
			newServiceAccount(),
			newRoute(r),
			newIngress(r),
			newTask(),
			newTaskRun(r),
			newRole(nil),
			newRoleBinding(nil),
			newPod(),
		}
	}
	return r.dependentTypes
}

func (r *ComponentManager) PrimaryResourceType() runtime.Object {
	return &component.Component{}
}

func (r *ComponentManager) NewFrom(name string, namespace string) (framework.Resource, error) {
	return controller2.NewComponent().FetchAndInit(name, namespace, r)
}

func (ComponentManager) asComponent(object runtime.Object) *controller2.Component {
	return object.(*controller2.Component)
}

func (r *ComponentManager) CreateOrUpdate(object framework.Resource) (err error) {
	c := r.asComponent(object)
	if component.BuildDeploymentMode == c.Spec.DeploymentMode {
		err = r.installBuildMode(c, c.Namespace)
	} else {
		err = r.installDevMode(c, c.Namespace)
	}
	return err
}

func (r *ComponentManager) Delete(resource framework.Resource) error {
	if r.IsTargetClusterRunningOpenShift() {
		// Delete the ImageStream created by OpenShift if it exists as the Component doesn't own this resource
		// when it is created during build deployment mode
		imageStream := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "image.openshift.io/v1",
				"kind":       "ImageStream",
				"metadata": map[string]interface{}{
					"name":      resource.GetName(),
					"namespace": resource.GetNamespace(),
				},
			},
		}

		// attempt to delete the imagestream if it exists
		if e := r.Client.Delete(context.TODO(), imageStream); e != nil && !errors.IsNotFound(e) {
			return e
		}
	}
	return nil
}
