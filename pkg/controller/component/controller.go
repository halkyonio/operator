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
	routev1 "github.com/openshift/api/route/v1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	controller2 "github.com/snowdrop/component-operator/pkg/controller"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// newReconciler returns a new reconcile.Reconciler
func NewComponentReconciler(mgr manager.Manager) *ReconcileComponent {
	// todo: make this configurable
	images := make(map[string]imageInfo, 7)
	defaultEnvVar := make(map[string]string, 7)
	// defaultEnvVar["JAVA_APP_JAR"] = "app.jar"
	images["spring-boot"] = imageInfo{
		registryRef: "quay.io/snowdrop/spring-boot-s2i",
		defaultEnv:  defaultEnvVar,
	}
	images["vert.x"] = imageInfo{
		registryRef: "quay.io/snowdrop/spring-boot-s2i",
		defaultEnv:  defaultEnvVar,
	}
	images["thorntail"] = imageInfo{
		registryRef: "quay.io/snowdrop/spring-boot-s2i",
		defaultEnv:  defaultEnvVar,
	}
	// References images
	images["openjdk8"] = imageInfo{registryRef: "registry.access.redhat.com/redhat-openjdk-18/openjdk18-openshift"}
	images["nodejs"] = imageInfo{registryRef: "nodeshift/centos7-s2i-nodejs"}
	images[supervisorImageId] = imageInfo{registryRef: "quay.io/snowdrop/supervisord"}

	supervisor := v1alpha2.Component{
		ObjectMeta: v1.ObjectMeta{
			Name: supervisorContainerName,
		},
		Spec: v1alpha2.ComponentSpec{
			Runtime: supervisorImageId,
			Version: latestVersionTag,
			Envs: []v1alpha2.Env{
				{
					Name: "CMDS",
					Value: "run-java:/usr/local/s2i/run;run-node:/usr/libexec/s2i/run;compile-java:/usr/local/s2i" +
						"/assemble;build:/deployments/buildapp",
				},
			},
		},
	}

	baseReconciler := controller2.NewBaseGenericReconciler(
		&v1alpha2.Component{},
		[]runtime.Object{
			&corev1.Pod{},
			&appsv1.Deployment{},
			&corev1.Service{},
			&routev1.Route{},
		}, mgr)
	r := &ReconcileComponent{
		BaseGenericReconciler: baseReconciler,
		runtimeImages:         images,
		supervisor:            &supervisor,
	}
	baseReconciler.SetReconcilerFactory(r)

	//r.initDependentResources()
	r.AddDependentResource(newPvc())
	r.AddDependentResource(newDeployment(r))
	r.AddDependentResource(newService(r))
	r.AddDependentResource(newServiceAccount())
	r.AddDependentResource(newRoute())
	r.AddDependentResource(newIngress())
	r.AddDependentResource(newTask())
	r.AddDependentResource(newTaskRun(r))

	return r
}

type imageInfo struct {
	registryRef string
	defaultEnv  map[string]string
}

type ReconcileComponent struct {
	*controller2.BaseGenericReconciler
	runtimeImages map[string]imageInfo
	supervisor    *v1alpha2.Component
	onOpenShift   *bool
}

func (r *ReconcileComponent) IsPrimaryResourceValid(object runtime.Object) bool {
	// todo: implement
	return true
}

func (ReconcileComponent) asComponent(object runtime.Object) *v1alpha2.Component {
	return object.(*v1alpha2.Component)
}

func (r *ReconcileComponent) ResourceMetadata(object runtime.Object) controller2.ResourceMetadata {
	component := r.asComponent(object)
	return controller2.ResourceMetadata{
		Name:         component.Name,
		Status:       component.Status.Phase.String(),
		Created:      component.ObjectMeta.CreationTimestamp,
		ShouldDelete: !component.ObjectMeta.DeletionTimestamp.IsZero(),
	}
}

func (r *ReconcileComponent) CreateOrUpdate(object runtime.Object) (bool, error) {
	component := r.asComponent(object)
	if v1alpha2.BuildDeploymentMode == component.Spec.DeploymentMode {
		return r.installBuildMode(component, component.Namespace)
	}
	return r.installDevMode(component, component.Namespace)
}

func (r *ReconcileComponent) SetErrorStatus(object runtime.Object, e error) {
	r.setErrorStatus(r.asComponent(object), e)
}

func (r *ReconcileComponent) SetSuccessStatus(object runtime.Object) {
	component := r.asComponent(object)
	if component.Status.Phase != v1alpha2.ComponentReady {
		err := r.updateStatus(component, v1alpha2.ComponentReady)
		if err != nil {
			panic(err)
		}
	}
}

func (r *ReconcileComponent) setInitialStatus(component *v1alpha2.Component, phase v1alpha2.ComponentPhase) error {
	if component.Generation == 1 && component.Status.Phase == "" {
		if err := r.updateStatus(component, phase); err != nil {
			r.ReqLogger.Info("Status update failed !")
			return err
		}
	}
	return nil
}
