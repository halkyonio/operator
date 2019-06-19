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
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	pipeline "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileComponent) installBuildMode(component *v1alpha2.Component, namespace string) (bool, error) {
	// Create Task s2i Buildah Push if it does not exists
	hasChanges := newFalse()
	if e := r.createAndCheckForChanges(component, &pipeline.Task{}, hasChanges); e != nil {
		return false, e
	}

	// Create ServiceAccount used by the Task's pod if it does not exists
	if e := r.createAndCheckForChanges(component, &v1.ServiceAccount{}, hasChanges); e != nil {
		return false, e
	}

	// Create the TaskRun in order to trigger the build
	if e := r.createAndCheckForChanges(component, &pipeline.TaskRun{}, hasChanges); e != nil {
		return false, e
	}

	if e := r.createDeploymentForBuildMode(component, hasChanges); e != nil {
		return false, e
	}

	if e := r.updateServiceSelector(component, hasChanges); e != nil {
		return false, e
	}

	return *hasChanges, nil
}

func (r *ReconcileComponent) createDeploymentForBuildMode(component *v1alpha2.Component, hasChanges *bool) error {

	// TODO : Review the logic maybe to check if the Deplpyment resource already exists when Deployment strategy = build

	// Create a new Deployment resource using the Deployment object to be used for a container to be created using a
	// container image
	obj, e := r.createBuildDeployment(component)
	if e != nil {
		return fmt.Errorf("deployment for the runtime container can't be created")
	}

	buildDeployment := obj.(*appsv1.Deployment)
	buildDeployment.Name = component.Name + "-build"
	buildDeployment.Namespace = component.Namespace
	controllerutil.SetControllerReference(component, buildDeployment, r.scheme)

	// We will check if a Dev Deployment exists.
	// If this is the case, then that means that we are switching from dev to build mode
	// and we will enrich the deployment resource of the runtime container
	devDeployment, e := r.fetchDeployment(component)
	if e == nil {
		devContainer := &devDeployment.Spec.Template.Spec.Containers[0]
		buildContainer := &buildDeployment.Spec.Template.Spec.Containers[0]
		buildContainer.Env = devContainer.Env
		buildContainer.EnvFrom = devContainer.EnvFrom
		buildContainer.Env = r.UpdateEnv(buildContainer.Env, component.Annotations["app.openshift.io/java-app-jar"])
	}

	// Create the Deployment object
	e = r.client.Create(context.TODO(), buildDeployment)
	if e != nil {
		return fmt.Errorf("Failed to create new deployment for the runtime container")
	}
	return nil
}

func (r *ReconcileComponent) UpdateEnv(envs []v1.EnvVar, jarName string) []v1.EnvVar {
	newEnvs := []v1.EnvVar{}
	for _, s := range envs {
		if s.Name == "JAVA_APP_JAR" {
			newEnvs = append(newEnvs, v1.EnvVar{Name: s.Name, Value: jarName})
		} else {
			newEnvs = append(newEnvs, s)
		}
	}
	return newEnvs
}

func (r *ReconcileComponent) updateServiceSelector(component *v1alpha2.Component, hasChanges *bool) error {
	componentName := component.Annotations["app.openshift.io/component-name"]
	svc := &v1.Service{}
	svc.Labels = map[string]string{
		"app.kubernetes.io/name":   componentName,
		"app.openshift.io/runtime": component.Spec.Runtime,
	}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: componentName, Namespace: component.Namespace}, svc); err != nil {
		return err
	}

	var nameApp string
	if v1alpha2.BuildDeploymentMode == component.Spec.DeploymentMode {
		nameApp = componentName + "-build"
	} else {
		nameApp = componentName
	}
	svc.Spec.Selector = map[string]string{
		"app": nameApp,
	}
	if err := r.client.Update(context.TODO(), svc); err != nil {
		return err
	}
	log.Info("### Updated Capability Selector to switch to a different component.")
	return nil
}
