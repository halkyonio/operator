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
	"fmt"
	component "halkyon.io/api/component/v1beta1"
	hLink "halkyon.io/api/link/v1beta1"
	"halkyon.io/api/v1beta1"
	controller2 "halkyon.io/operator/pkg/controller"
	"halkyon.io/operator/pkg/controller/framework"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

// newReconciler returns a new reconcile.Reconciler
func NewComponentReconciler() *ReconcileComponent {
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

	r := &ReconcileComponent{
		runtimeImages: images,
		supervisor:    &supervisor,
	}
	return r
}

type imageInfo struct {
	registryRef string
	defaultEnv  map[string]string
}

type ReconcileComponent struct {
	*framework.K8SHelper
	runtimeImages map[string]imageInfo
	supervisor    *component.Component
}

func (r *ReconcileComponent) SetHelper(helper *framework.K8SHelper) {
	r.K8SHelper = helper
}

func (r *ReconcileComponent) Helper() *framework.K8SHelper {
	return r.K8SHelper
}

func (r *ReconcileComponent) GetDependentResourcesTypes() []framework.DependentResource {
	return []framework.DependentResource{
		newPvc(),
		newDeployment(r),
		newService(r),
		newServiceAccount(),
		newRoute(r),
		newIngress(r),
		newTask(),
		newTaskRun(r),
		controller2.NewRole(),
		controller2.NewRoleBinding(),
		newPod(),
	}
}

func (r *ReconcileComponent) PrimaryResourceType() runtime.Object {
	return &component.Component{}
}

func (r *ReconcileComponent) NewFrom(name string, namespace string) (framework.Resource, error) {
	c := controller2.NewComponent()
	_, err := r.Fetch(name, namespace, c.Component)
	resourcesTypes := r.GetDependentResourcesTypes()
	for _, rType := range resourcesTypes {
		c.AddDependentResource(rType.NewInstanceWith(c))
	}
	return c, err
}

func (ReconcileComponent) asComponent(object runtime.Object) *controller2.Component {
	return object.(*controller2.Component)
}

func (r *ReconcileComponent) CreateOrUpdate(object framework.Resource) (err error) {
	c := r.asComponent(object)
	if component.BuildDeploymentMode == c.Spec.DeploymentMode {
		err = r.installBuildMode(c, c.Namespace)
	} else {
		err = r.installDevMode(c, c.Namespace)
	}
	return err
}

func (r *ReconcileComponent) Delete(resource framework.Resource) error {
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

func (r *ReconcileComponent) SetPrimaryResourceStatus(primary framework.Resource, statuses []framework.DependentResourceStatus) (needsUpdate bool) {
	c := r.asComponent(primary)
	if len(c.Status.Links) > 0 {
		for i, link := range c.Status.Links {
			if link.Status == component.Started {
				p, err := c.FetchUpdatedDependent(&corev1.Pod{}, r.K8SHelper)
				name := p.(*corev1.Pod).Name
				if err != nil || name == link.OriginalPodName {
					c.Status.Phase = component.ComponentLinking
					c.SetNeedsRequeue(true)
					return false
				} else {
					// update link status
					l := &hLink.Link{}
					err := r.Client.Get(context.TODO(), types.NamespacedName{
						Namespace: c.Namespace,
						Name:      link.Name,
					}, l)
					if err != nil {
						// todo: is this appropriate?
						link.Status = component.Errored
						c.Status.Message = fmt.Sprintf("couldn't retrieve '%s' link", link.Name)
						return true
					}

					l.Status.Message = fmt.Sprintf("'%s' finished linking", c.Name)
					err = r.Client.Status().Update(context.TODO(), l)
					if err != nil {
						// todo: fix-me
						r.ReqLogger.Error(err, "couldn't update link status", "link name", l.Name)
					}

					link.Status = component.Linked
					link.OriginalPodName = ""
					c.Status.PodName = name
					c.Status.Links[i] = link // make sure we update the links with the modified value
					needsUpdate = true
				}
			}
		}
	}
	// make sure we propagate the need for update even if setting the status doesn't change anything
	return primary.SetSuccessStatus(statuses, "Ready") || needsUpdate
}
