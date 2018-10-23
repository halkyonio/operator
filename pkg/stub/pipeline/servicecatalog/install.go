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

package servicecatalog

import (
	"encoding/json"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/operator-framework/operator-sdk/pkg/util/k8sutil"
	log "github.com/sirupsen/logrus"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha1"
	"github.com/snowdrop/component-operator/pkg/stub/pipeline"
	"github.com/snowdrop/component-operator/pkg/util/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servicecatalog "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	util "github.com/snowdrop/component-operator/pkg/util/template"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"strings"

	// metav1unstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"text/template"
)

// NewServiceInstanceStep creates a step that handles the creation of the Service from the catalog
func NewServiceInstanceStep() pipeline.Step {
	return &serviceInstanceStep{}
}

type serviceInstanceStep struct {
}

func (serviceInstanceStep) Name() string {
	return "create service"
}

func (serviceInstanceStep) CanHandle(component *v1alpha1.Component) bool {
	return component.Status.Phase == ""
}

func (serviceInstanceStep) Handle(component *v1alpha1.Component, deleted bool) error {
	target := component.DeepCopy()
	if (deleted) {
		return deleteService(target)
	} else {
		return createService(target)
	}
}

func deleteService(component *v1alpha1.Component) error {
	for _, s := range component.Spec.Services {
		// Let's retrieve the ServiceBindings to delete them first
		listServiceBindings := new(servicecatalog.ServiceBindingList)
		listServiceBindings.TypeMeta = metav1.TypeMeta{
			Kind:       "ServiceBinding",
			APIVersion: "servicecatalog.k8s.io/v1beta1",
		}
		err := sdk.List(component.ObjectMeta.Namespace, listServiceBindings)
		if err != nil {
			return err
		}
		// Delete ServiceBinding(s) linked to the ServiceInstance
		for _, sb := range listServiceBindings.Items {
			if sb.Name == s.Name {
				err := sdk.Delete(&sb)
				if err != nil {
					return err
				}
				log.Infof("ServiceBinding %s deleted",sb.Name)
			}
		}

		// Retrieve ServiceInstances
		listServiceInstance := new(servicecatalog.ServiceInstanceList)
		listServiceInstance.TypeMeta = metav1.TypeMeta{
			Kind:       "ServiceInstance",
			APIVersion: "servicecatalog.k8s.io/v1beta1",
		}
		err = sdk.List(component.ObjectMeta.Namespace, listServiceInstance)
		if err != nil {
			return err
		}

		// Delete ServiceInstance(s)
		for _, si := range listServiceInstance.Items {
			err := sdk.Delete(&si)
			if err != nil {
				return err
			}
			log.Infof("Service Instance %s deleted",si.Name)
		}
	}
	return nil
}

func createService(component *v1alpha1.Component) error {
	// Get Current Namespace
	namespace, err := kubernetes.GetClientCurrentNamespace("")
	if err != nil {
		return err
	}
	component.ObjectMeta.Namespace = namespace

	for i, s := range component.Spec.Services {
		// Convert the parameters into a JSon string
		mapParams := ParametersAsMap(s.Parameters)
		rawJSON := string(BuildParameters(mapParams).Raw)
		component.Spec.Services[i].ParametersJSon = rawJSON

		// Create the ServiceInstance and ServiceBinding using the template
		for _, tmpl := range util.Templates {
			if strings.HasPrefix(tmpl.Name(), "servicecatalog") {
				err := createResource(tmpl, component)
				if err != nil {
					return err
				}
			}
		}
	}

	log.Infof("%s service created", component.Name)
	component.Status.Phase = v1alpha1.PhaseServiceCreation

	err = sdk.Update(component)
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

// BuildParameters converts a map of variable assignments to a byte encoded json document,
// which is what the ServiceCatalog API consumes.
func BuildParameters(params interface{}) *runtime.RawExtension {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		// This should never be hit because marshalling a map[string]string is pretty safe
		// I'd rather throw a panic then force handling of an error that I don't think is possible.
		log.Errorf("unable to marshal the request parameters %v (%s)", params, err)
	}
	return &runtime.RawExtension{Raw: paramsJSON}
}

// Convert Array of parameters to a Map
func ParametersAsMap(parameters []v1alpha1.Parameter) map[string]string {
	result := make(map[string]string)
	for _, parameter := range parameters {
		result[parameter.Name] = parameter.Value
	}
	return result
}

func createResource(tmpl template.Template, component *v1alpha1.Component) error {
	res, err := newResourceFromTemplate(tmpl, component)
	if err != nil {
		return err
	}

	for _, r := range res {
		err = sdk.Create(r)
		if err != nil && !k8serrors.IsAlreadyExists(err) {
			return err
		}
	}

	return nil
}

func newResourceFromTemplate(template template.Template, component *v1alpha1.Component) ([]runtime.Object, error) {
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
			obj, err := k8sutil.RuntimeObjectFromUnstructured(&item)
			if err != nil {
				return nil, err
			}
			kind := obj.GetObjectKind().GroupVersionKind().Kind
			if strings.HasPrefix(kind, "ServiceInstance") || strings.HasPrefix(kind, "ServiceBinding") {
				kubernetes.SetNamespace(obj, component)
			} else {
				kubernetes.SetNamespaceAndOwnerReference(obj, component)
			}
			result = append(result, obj)
		}
	} else {
		obj, err := k8sutil.RuntimeObjectFromUnstructured(r)
		if err != nil {
			return nil, err
		}

		kind := obj.GetObjectKind().GroupVersionKind().Kind
		if strings.HasPrefix(kind, "ServiceInstance") || strings.HasPrefix(kind, "ServiceBinding") {
			kubernetes.SetNamespace(obj, component)
		} else {
			kubernetes.SetNamespaceAndOwnerReference(obj, component)
		}
		result = append(result, obj)
	}
	return result, nil
}
