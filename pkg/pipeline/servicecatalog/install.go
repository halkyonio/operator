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
	"context"
	"encoding/json"
	servicecatalog "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	log "github.com/sirupsen/logrus"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha1"
	"github.com/snowdrop/component-operator/pkg/pipeline"
	. "github.com/snowdrop/component-operator/pkg/util/helper"
	"github.com/snowdrop/component-operator/pkg/util/kubernetes"
	util "github.com/snowdrop/component-operator/pkg/util/template"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"

	// metav1unstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"text/template"
)

// NewServiceInstanceStep creates a step that handles the creation of the Service from the catalog
func NewServiceInstanceStep() pipeline.Step {
	return &newServiceInstanceStep{}
}

type newServiceInstanceStep struct {
}

func (newServiceInstanceStep) Name() string {
	return "create service"
}

// Service is installed when the status of the component is empty.
// Such case occurs the first time the component is created AND before the innerloop takes place
func (newServiceInstanceStep) CanHandle(component *v1alpha1.Component) bool {
	return component.Status.Phase == "" || component.Status.Phase == v1alpha1.PhaseDeploying
}

func (newServiceInstanceStep) Handle(component *v1alpha1.Component, client *client.Client, namespace string) error {
	return createService(component, *client, namespace)
}

func createService(component *v1alpha1.Component, c client.Client, namespace string) error {
	component.ObjectMeta.Namespace = namespace

	for i, s := range component.Spec.Services {
		// Convert the parameters into a JSon string
		mapParams := ParametersAsMap(s.Parameters)
		rawJSON := string(BuildParameters(mapParams).Raw)
		component.Spec.Services[i].ParametersJSon = rawJSON

		// Create the ServiceInstance and ServiceBinding using the template
		for _, tmpl := range util.Templates {
			if strings.HasPrefix(tmpl.Name(), "servicecatalog") {
				err := createResource(tmpl, component, c)
				if err != nil {
					return err
				}
			}
		}
		log.Infof("#### Created service instance's '%s' for the service/class '%s' and plan '%s'", s.Name, s.Class, s.Plan)
	}

	log.Infof("#### Created %s CRD's service component", component.Name)
	component.Status.Phase = v1alpha1.PhaseServiceCreation
	svcFinalizerName := "service.component.k8s.io"
	if !ContainsString(component.ObjectMeta.Finalizers, svcFinalizerName) {
		component.ObjectMeta.Finalizers = append(component.ObjectMeta.Finalizers, svcFinalizerName)
	}

	// err := c.Update(context.TODO(), component)
	err := c.Status().Update(context.TODO(), component)
	if err != nil && k8serrors.IsConflict(err) {
		return err
	}
	log.Info("### Pipeline 'service catalog' ended ###")
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

func createResource(tmpl template.Template, component *v1alpha1.Component, c client.Client) error {
	res, err := newResourceFromTemplate(tmpl, component)
	if err != nil {
		return err
	}

	for _, r := range res {
		err = c.Create(context.TODO(), r)
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
			obj, err := kubernetes.RuntimeObjectFromUnstructured(&item)
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
		obj, err := kubernetes.RuntimeObjectFromUnstructured(r)
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

func listServiceBindings(component *v1alpha1.Component, c client.Client) (*servicecatalog.ServiceBindingList, error) {
	listServiceBinding := new(servicecatalog.ServiceBindingList)
	listServiceBinding.TypeMeta = metav1.TypeMeta{
		Kind:       "ServiceBinding",
		APIVersion: "servicecatalog.k8s.io/v1beta1",
	}
	listOps := client.ListOptions{
		Namespace:     component.ObjectMeta.Namespace,
		LabelSelector: getLabelsSelector(component.ObjectMeta.Labels),
	}
	err := c.List(context.TODO(), &listOps, listServiceBinding)
	if err != nil {
		return nil, err
	}
	return listServiceBinding, nil
}

func getLabelsSelector(mapLabels map[string]string) labels.Selector {
	return labels.SelectorFromSet(mapLabels)
}

func getComponentSelector() metav1.ListOptions {
	return metav1.ListOptions{
		LabelSelector: "app=my-spring-boot-service",
	}
}
