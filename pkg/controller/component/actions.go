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
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"golang.org/x/net/context"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/snowdrop/component-operator/pkg/util/kubernetes"
	util "github.com/snowdrop/component-operator/pkg/util/template"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"text/template"
)

func (r *ReconcileComponent) installInnerLoop(component *v1alpha2.Component, namespace string) error {
	component.ObjectMeta.Namespace = namespace
	// Append dev runtime's image (java, nodejs, ...)
	component.Spec.RuntimeName = strings.Join([]string{"dev-runtime", strings.ToLower(component.Spec.Runtime)}, "-")
	// Enrich Component with k8s recommend Labels
	component.ObjectMeta.Labels = kubernetes.PopulateK8sLabels(component, "Backend")

	isOpenshift, err := kubernetes.DetectOpenShift(r.config)
	if err != nil {
		return err
	}

	if (isOpenshift) {

		// Create ImageStream if it does not exists
		imageStreamToCreate := []string{}
		for _, name := range r.getDevImageNames(component) {
			if _, err := r.fetchImageStream(component, name); err != nil {
				imageStreamToCreate = append(imageStreamToCreate, name)
			}
		}

		for _, name := range imageStreamToCreate {
			if err := r.client.Create(context.TODO(),r.buildImageStream(component, name)); err != nil {
				if err != nil {
					return err
				}
			}
			r.reqLogger.Info(fmt.Sprintf("Created imagestream : %s",name))
		}


		tmpl, ok := util.Templates["innerloop/deploymentconfig"]
		if ok {
			if component.Spec.Port == 0 {
				component.Spec.Port = 8080 // Add a default port if empty
			}
			component.Spec.SupervisordName = "copy-supervisord"

			// Enrich Env Vars with Default values
			r.populateEnvVar(component)

			// Create DeploymentConfig if it does not exists
			if _, err := r.fetchDeploymentConfig(component); err != nil {
				err := CreateResource(tmpl, component, r.client, r.scheme)
				if err != nil {
					return err
				}
				r.reqLogger.Info("Created deployment config")
			}
		}

		if component.Spec.ExposeService {
			// Create Route if it does not exists
			if _, err := r.fetchRoute(component); err != nil {
				if _, err := r.create(component, ROUTE, err); err != nil {
					if err != nil {
						return err
					}
				}
				r.reqLogger.Info("Create route", "Spec port", component.Spec.Port)
			}
		}
	} else {
		// This is not an OpenShift cluster but instead a K8s platform
		if component.Spec.Port == 0 {
			component.Spec.Port = 8080 // Add a default port if empty
		}
		component.Spec.SupervisordName = "copy-supervisord"
		// Enrich Env Vars with Default values
		r.populateEnvVar(component)

		// Create Deployment if it does not exists
		if _, err := r.fetchDeployment(component); err != nil {
			if _, err := r.create(component, DEPLOYMENT, err); err != nil {
				return err
			} else {
				r.reqLogger.Info("Created deployment")
			}
		}

		if component.Spec.ExposeService {
			if _, err := r.fetchRoute(component); err != nil {
				if _, err := r.create(component, INGRESS, err); err != nil {
					if err != nil {
						return err
					}
				}
				r.reqLogger.Info("Created ingress", "Port", component.Spec.Port)
			}
		}
	}

	// Install common resources

	// Create PVC if it does not exists
	component.Spec.Storage.Capacity = "1Gi"
	component.Spec.Storage.Mode = "ReadWriteOnce"
	component.Spec.Storage.Name = "m2-data-" + component.Name
	if _, err := r.fetchPVC(component); err != nil {
		if _, err := r.create(component, PERSISTENTVOLUMECLAIM, err); err != nil {
			if err != nil {
				return err
			}
		}
		r.reqLogger.Info("Created pvc", "Name", component.Spec.Storage.Name, "Capacity", component.Spec.Storage.Capacity, "Mode", component.Spec.Storage.Mode)

	}

	// Create Service if it does not exists
	if component.Spec.Port == 0 {
		component.Spec.Port = 8080 // Add a default port if empty
	}
	if _, err := r.fetchService(component); err != nil {
		if _, err := r.create(component, SERVICE, err); err != nil {
			if err != nil {
				return err
			}
		}
		r.reqLogger.Info("Created service", "Spec port", component.Spec.Port)
	}

	r.reqLogger.Info("Deploying Component")
	return nil
}

func (r *ReconcileComponent) installIOuterLoop(component *v1alpha2.Component, namespace string) error {
	return nil
}

func CreateResource(tmpl template.Template, component *v1alpha2.Component, c client.Client, scheme *runtime.Scheme) error {
	res, err := newResourceFromTemplate(tmpl, component, scheme)
	if err != nil {
		return err
	}

	for _, r := range res {
		if obj, ok := r.(metav1.Object); ok {
			obj.SetLabels(kubernetes.PopulateK8sLabels(component, "Backend"))
		}
		err = c.Create(context.TODO(), r)
		if err != nil && k8serrors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

func newResourceFromTemplate(template template.Template, component *v1alpha2.Component, scheme *runtime.Scheme) ([]runtime.Object, error) {
	var result = []runtime.Object{}

	var b = util.Parse(template, component)
	r, err := kubernetes.PopulateKubernetesObjectFromYaml(b.String())
	if err != nil {
		return nil, err
	}

	if strings.HasSuffix(r.GetKind(), "List") {
		l, err := r.ToList()
		if err != nil {
			return nil, err
		}
		for _, item := range l.Items {
			obj, err := kubernetes.RuntimeObjectFromUnstructured(&item)
			if err != nil {
				return nil, err
			}
			ro, ok := obj.(v1.Object)
			ro.SetNamespace(component.Namespace)
			if !ok {
				return nil, err
			}
			controllerutil.SetControllerReference(component, ro, scheme)
			//kubernetes.SetNamespaceAndOwnerReference(obj, component)
			result = append(result, obj)
		}
	} else {
		obj, err := kubernetes.RuntimeObjectFromUnstructured(r)
		if err != nil {
			return nil, err
		}

		ro, ok := obj.(v1.Object)
		ro.SetNamespace(component.Namespace)
		if !ok {
			return nil, err
		}
		controllerutil.SetControllerReference(component, ro, scheme)
		//kubernetes.SetNamespaceAndOwnerReference(obj, component)
		result = append(result, obj)
	}
	return result, nil
}
