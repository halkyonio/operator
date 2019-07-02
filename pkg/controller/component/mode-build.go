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
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
)

func (r *ReconcileComponent) installBuildMode(component *v1alpha2.Component, namespace string) (changed bool, e error) {
	// Create Task s2i Buildah Push if it does not exists
	if changed, e = r.CreateIfNeeded(component, &pipeline.Task{}); e != nil {
		return false, e
	}

	// Create ServiceAccount used by the Task's pod if it does not exists
	if changed, e = r.CreateIfNeeded(component, &corev1.ServiceAccount{}); e != nil {
		return false, e
	}

	// Create the TaskRun in order to trigger the build
	if changed, e = r.CreateIfNeeded(component, &pipeline.TaskRun{}); e != nil {
		return false, e
	}

	if changed, e = r.CreateIfNeeded(component, &v1.Deployment{}); e != nil {
		return false, e
	}

	if changed, e = r.CreateIfNeeded(component, &corev1.Service{}); e != nil {
		return false, e
	}

	return changed, nil
}

func (r *ReconcileComponent) updateServiceSelector(object runtime.Object, res dependentResource, component *v1alpha2.Component) (bool, error) {
	svc, ok := object.(*corev1.Service)
	if !ok {
		return false, fmt.Errorf("updateServiceSelector only works on Service instances, got '%s'", reflect.TypeOf(object).Elem().Name())
	}

	// update the service selector if needed
	name := res.labelsName(component)
	if svc.Spec.Selector["app"] != name {
		svc.Spec.Selector["app"] = name
		if err := r.Client.Update(context.TODO(), svc); err != nil {
			return false, fmt.Errorf("couldn't update service '%s' selector", svc.Name)
		}
		return true, nil
	}

	return false, nil
}
