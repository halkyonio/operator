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
	"fmt"
	halkyon "halkyon.io/operator/pkg/controller"
)

func (r *ComponentManager) installDevMode(component *halkyon.Component, namespace string) (e error) {
	component.ObjectMeta.Namespace = namespace
	// Enrich Component with k8s recommend Labels
	component.ObjectMeta.Labels = r.PopulateK8sLabels(component, "Backend")
	// Check if Service port exists, otherwise error out
	if component.Spec.Port == 0 {
		return fmt.Errorf("component '%s' must provide a port", component.Name)
	}

	// Enrich Env Vars with Default values
	populateEnvVar(component)

	return component.CreateOrUpdate(r.Helper())
}
