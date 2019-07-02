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
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	pipeline "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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
