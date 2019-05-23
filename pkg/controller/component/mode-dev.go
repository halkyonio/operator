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
	// v1 "k8s.io/api/core/v1"
	// "k8s.io/apimachinery/pkg/runtime"
	// "k8s.io/apimachinery/pkg/types"
	// "k8s.io/client-go/rest"
	// "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/snowdrop/component-operator/pkg/util"
)

func (r *ReconcileComponent) installDevMode(component *v1alpha2.Component, namespace string) error {
	component.ObjectMeta.Namespace = namespace
	// Enrich Component with k8s recommend Labels
	component.ObjectMeta.Labels = r.PopulateK8sLabels(component, "Backend")
	// Check if Capability port exists, otherwise define it
	if component.Spec.Port == 0 {
		component.Spec.Port = 8080 // Add a default port if empty
	}

	// Specify the default Storage data - value
	component.Spec.Storage.Capacity = "1Gi"
	component.Spec.Storage.Mode = "ReadWriteOnce"
	component.Spec.Storage.Name = "m2-data-" + component.Name

	// Enrich Env Vars with Default values
	r.populateEnvVar(component)

	isOpenShift, err := util.IsOpenshift(r.config)
	if err != nil {
		return err
	}

	// Install common resources

	// Create PVC if it does not exists
	if _, err := r.fetchPVC(component); err != nil {
		if _, err := r.create(component, PERSISTENTVOLUMECLAIM, err); err != nil {
			return err
		}
		r.reqLogger.Info("Created pvc", "Name", component.Spec.Storage.Name, "Capacity", component.Spec.Storage.Capacity, "Mode", component.Spec.Storage.Mode)

	}

	// Create Deployment if it does not exists
	if _, err := r.fetchDeployment(component); err != nil {
		if _, err := r.create(component, DEPLOYMENT, err); err != nil {
			return err
		} else {
			r.reqLogger.Info("Created deployment")
		}
	}

	if _, err := r.fetchService(component); err != nil {
		if _, err := r.create(component, SERVICE, err); err != nil {
			return err
		}
		r.reqLogger.Info("Created service", "Spec port", component.Spec.Port)
	}

	if component.Spec.ExposeService {
		if isOpenShift {
			// Create an OpenShift Route
			if _, err := r.fetchRoute(component); err != nil {
				if _, err := r.create(component, ROUTE, err); err != nil {
					return err
				}
				r.reqLogger.Info("Create route", "Spec port", component.Spec.Port)
			}
		} else {
			// Create an Ingress resource
			if _, err := r.fetchRoute(component); err != nil {
				if _, err := r.create(component, INGRESS, err); err != nil {
					return err
				}
				r.reqLogger.Info("Created ingress", "Port", component.Spec.Port)
			}
		}
	}

	r.reqLogger.Info("Deploying Component")
	return nil
}
