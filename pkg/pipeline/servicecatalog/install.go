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
	"golang.org/x/net/context"
	"encoding/json"
	servicecatalog "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	log "github.com/sirupsen/logrus"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"github.com/snowdrop/component-operator/pkg/pipeline"
	. "github.com/snowdrop/component-operator/pkg/util/helper"
	"github.com/snowdrop/component-operator/pkg/util/kubernetes"
	util "github.com/snowdrop/component-operator/pkg/util/template"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
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
func (newServiceInstanceStep) CanHandle(component *v1alpha2.Component) bool {
	// log.Infof("## Status to be checked : %s", component.Status.Phase)
	return component.Status.Phase == "" || component.Status.Phase == v1alpha2.PhaseDeploying
}

func (newServiceInstanceStep) Handle(component *v1alpha2.Component, config *rest.Config, client *client.Client, namespace string, scheme *runtime.Scheme) error {
	return createService(*component, *config, *client, namespace, *scheme)
}

func createService(component v1alpha2.Component, config rest.Config, c client.Client, namespace string, scheme runtime.Scheme) error {
	component.ObjectMeta.Namespace = namespace

	IsServiceInstalled := false

	// IF the ServiceInstance/ServiceBinding already exists, then we try only to change the status
	// Let's retrieve the ServiceBindings to delete them first
	list, err := listServiceBindings(&component, c)
	if err != nil {
		return err
	}

	// If the ServiceBinding List is not empty, then we have a binding with a Secret and ServiceInstance
	if len(list.Items) > 0 {
		// Retrieve ServiceInstances
		serviceInstanceList := new(servicecatalog.ServiceInstanceList)
		serviceInstanceList.TypeMeta = metav1.TypeMeta{
			Kind:       "ServiceInstance",
			APIVersion: "servicecatalog.k8s.io/v1beta1",
		}
		listOps := &client.ListOptions{
			Namespace: component.ObjectMeta.Namespace,
		}
		err = c.List(context.TODO(), listOps, serviceInstanceList)
		if err != nil {
			return err
		}

		if len(serviceInstanceList.Items) > 0 {
			IsServiceInstalled = true
		}
	}

	if IsServiceInstalled != true {
		for i, s := range component.Spec.Services {
			// Convert the parameters into a JSon string
			mapParams := ParametersAsMap(s.Parameters)
			rawJSON := string(BuildParameters(mapParams).Raw)
			component.Spec.Services[i].ParametersJSon = rawJSON

			// Create the ServiceInstance and ServiceBinding using the template
			tmpl, ok := util.Templates["servicecatalog/serviceinstance"]
			if ok {
				err := createResource(tmpl, &component, c)
				if err != nil {
					log.Infof("## Service instance creation failed !")
					return err
				}
			}
			tmpl, ok = util.Templates["servicecatalog/servicebinding"]
			if ok {
				err := createResource(tmpl, &component, c)
				if err != nil {
					log.Infof("## Service Binding creation failed !")
					return err
				}
			}
			log.Infof("### Created service instance's '%s' for the service/class '%s' and plan '%s'", s.Name, s.Class, s.Plan)
		}

		log.Infof("### Created %s CRD's service component", component.Name)
	}

	// Fetch the latest Component created (as it could have been modified by another step of the pipeline)
	newComponent := &v1alpha2.Component{}
	err = c.Get(context.TODO(), types.NamespacedName{Name: component.Name, Namespace: component.Namespace}, newComponent)
	if err != nil {
		return err
	}
	// Update status
	newComponent.Status.Phase = v1alpha2.PhaseServiceCreation
	// Add Finalizer to allow the serviceinstance/binding to be deleted when the component will be deleted
	svcFinalizerName := "service.component.k8s.io"
	if !ContainsString(newComponent.ObjectMeta.Finalizers, svcFinalizerName) {
		newComponent.ObjectMeta.Finalizers = append(newComponent.ObjectMeta.Finalizers, svcFinalizerName)
	}
	err = c.Update(context.TODO(), newComponent)
	if err != nil && k8serrors.IsConflict(err) {
		log.Infof("## Component Service - status update failed")
		return err
	}
	log.Info("## Pipeline 'service catalog' ended ##")
	log.Infof("## Status updated : %s ##",newComponent.Status.Phase)
	log.Infof("## Status RevNumber : %s ##",newComponent.Status.RevNumber)
	log.Info("------------------------------------------------------")
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
func ParametersAsMap(parameters []v1alpha2.Parameter) map[string]string {
	result := make(map[string]string)
	for _, parameter := range parameters {
		result[parameter.Name] = parameter.Value
	}
	return result
}

func createResource(tmpl template.Template, component *v1alpha2.Component, c client.Client) error {
	res, err := newResourceFromTemplate(tmpl, component)
	if err != nil {
		return err
	}

	for _, r := range res {
		if obj, ok := r.(metav1.Object); ok {
			obj.SetLabels(kubernetes.PopulateK8sLabels(component,"Service"))
		}
		err = c.Create(context.TODO(), r)
		if err != nil && !k8serrors.IsAlreadyExists(err) {
			return err
		}
	}

	return nil
}

func newResourceFromTemplate(template template.Template, component *v1alpha2.Component) ([]runtime.Object, error) {
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

func listServiceBindings(component *v1alpha2.Component, c client.Client) (*servicecatalog.ServiceBindingList, error) {
	listServiceBinding := new(servicecatalog.ServiceBindingList)
	listServiceBinding.TypeMeta = metav1.TypeMeta{
		Kind:       "ServiceBinding",
		APIVersion: "servicecatalog.k8s.io/v1beta1",
	}
	listOps := client.ListOptions{
		Namespace:     component.ObjectMeta.Namespace,
		// LabelSelector: getLabelsSelector(component.ObjectMeta.Labels),
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
