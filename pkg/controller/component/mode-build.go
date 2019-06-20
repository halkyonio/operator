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
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	pipeline "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func (r *ReconcileComponent) installBuildMode(component *v1alpha2.Component, namespace string) (bool, error) {
	// Create Task s2i Buildah Push if it does not exists
	hasChanges := newFalse()
	if e := r.createAndCheckForChanges(component, &pipeline.Task{}, hasChanges); e != nil {
		return false, e
	}

	// Create ServiceAccount used by the Task's pod if it does not exists
	if e := r.createAndCheckForChanges(component, &corev1.ServiceAccount{}, hasChanges); e != nil {
		return false, e
	}

	// Create the TaskRun in order to trigger the build
	if e := r.createAndCheckForChanges(component, &pipeline.TaskRun{}, hasChanges); e != nil {
		return false, e
	}

	if e := r.createAndCheckForChanges(component, &v1.Deployment{}, hasChanges); e != nil {
		return false, e
	}

	if e := r.updateServiceSelector(component, hasChanges); e != nil {
		return false, e
	}

	return *hasChanges, nil
}

func (r *ReconcileComponent) updateServiceSelector(component *v1alpha2.Component, hasChanges *bool) error {

	var nameApp string

	if v1alpha2.BuildDeploymentMode == component.Spec.DeploymentMode {
		nameApp = component.Name + "-build"
	} else {
		nameApp = component.Name
	}

	if svc, e := r.fetchService(component); e != nil {
		// Service don't exist. So will create it
		obj, e := r.buildService(dependentResource{prototype: &corev1.Service{}, name: defaultNamer}, component)
		if e != nil {
			svc := obj.(*corev1.Service)
			svc.Spec.Selector = map[string]string{
				"app": nameApp,
			}
		} else {
			return e
		}
		if err := r.client.Create(context.TODO(), svc); err != nil {
			return err
		}
	} else {
		svc.Spec.Selector = map[string]string{
			"app": nameApp,
		}
		if err := r.client.Update(context.TODO(), svc); err != nil {
			return err
		}
	}
	*hasChanges = true
	return nil
}
