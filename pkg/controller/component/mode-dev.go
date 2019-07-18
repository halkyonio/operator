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
	routev1 "github.com/openshift/api/route/v1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/client-go/discovery"
	"strings"
)

func newTrue() *bool {
	b := true
	return &b
}

func newFalse() *bool {
	b := false
	return &b
}

func (r *ReconcileComponent) isTargetClusterRunningOpenShift() bool {
	if r.onOpenShift == nil {
		discoveryClient, err := discovery.NewDiscoveryClientForConfig(r.Config)
		if err != nil {
			panic(err)
		}
		apiList, err := discoveryClient.ServerGroups()
		if err != nil {
			panic(err)
		}
		apiGroups := apiList.Groups
		for _, group := range apiGroups {
			if strings.HasSuffix(group.Name, "openshift.io") {
				r.onOpenShift = newTrue()
				break
			}
		}

		if r.onOpenShift == nil {
			// we didn't find any api group with the openshift.io suffix, so we're not on OpenShift!
			r.onOpenShift = newFalse()
		}
	}

	return *r.onOpenShift
}

func (r *ReconcileComponent) installDevMode(component *v1alpha2.Component, namespace string) (e error) {
	component.ObjectMeta.Namespace = namespace
	// Enrich Component with k8s recommend Labels
	component.ObjectMeta.Labels = r.PopulateK8sLabels(component, "Backend")
	// Check if Service port exists, otherwise error out
	if component.Spec.Port == 0 {
		return fmt.Errorf("component '%s' must provide a port", component.Name)
	}

	// Enrich Env Vars with Default values
	r.populateEnvVar(component)

	// Create PVC if it does not exists
	if e = r.CreateIfNeeded(component, &corev1.PersistentVolumeClaim{}); e != nil {
		return e
	}

	// Create Deployment if it does not exists
	if e = r.CreateIfNeeded(component, &appsv1.Deployment{}); e != nil {
		return e
	}

	if e = r.CreateIfNeeded(component, &corev1.Service{}); e != nil {
		return e
	}

	if component.Spec.ExposeService {
		if r.isTargetClusterRunningOpenShift() {
			// Create an OpenShift Route
			if e = r.CreateIfNeeded(component, &routev1.Route{}); e != nil {
				return e
			}
		} else {
			// Create an Ingress resource
			if e = r.CreateIfNeeded(component, &v1beta1.Ingress{}); e != nil {
				return e
			}
		}
	}

	return
}

func (r *ReconcileComponent) deleteDevMode(component *v1alpha2.Component, namespace string) error {
	// todo
	return nil
}
