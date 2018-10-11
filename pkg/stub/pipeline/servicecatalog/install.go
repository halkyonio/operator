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
	util "github.com/snowdrop/component-operator/pkg/util/template"
	appsv1 "github.com/openshift/api/apps/v1"
	deploymentconfig "github.com/openshift/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
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

func (serviceInstanceStep) Handle(component *v1alpha1.Component) error {
	target := component.DeepCopy()
	return createService(target)
}

func createService(component *v1alpha1.Component) error {
	// Get Current Namespace
	namespace, err := kubernetes.GetClientCurrentNamespace("")
	if err != nil {
		return err
	}
	component.ObjectMeta.Namespace = namespace

	// Convert the parameters into a JSon string
	newServices := []v1alpha1.Service{}
	for _, s := range component.Spec.Services {
		mapParams := ParametersAsMap(s.Parameters)
		rawJSON := string(BuildParameters(mapParams).Raw)
		s.ParametersJSon = rawJSON
		newServices = append(newServices,s)
	}
	component.Spec.Services = newServices

	// Create the ServiceInstance and ServiceBinding using the template
	for _, tmpl := range util.Templates {
		if strings.HasPrefix(tmpl.Name(), "servicecatalog") {
			err := createResource(tmpl, component)
			if err != nil {
				return err
			}
		}
	}

	// Get DeploymentConfig to inject EnvFrom using Secret and restart it
	dc, err := GetDeploymentConfig(namespace, component.Name)
	if err != nil {
		return err
	}

	// TODO
	secretName := "my-postgresql-db"
	// Add the Secret as EnvVar to the container
	dc.Spec.Template.Spec.Containers[0].EnvFrom = addSecretAsEnvFromSource(secretName)

	// Update the DeploymentConfig
	err = sdk.Update(dc)
	if err != nil {
		log.Fatalf("DeploymentConfig not updated : %s", err.Error())
	}
	log.Infof("'%s' EnvFrom secret added to the DeploymentConfig", secretName)

	// Create a DeploymentRequest and redeploy it
	request := &appsv1.DeploymentRequest{
		Name:   component.Name,
		Latest: true,
		Force:  true,
	}

	err = sdk.Update(request)
	if err != nil {
		log.Fatalf("Redeployment of the DeploymentConfig failed %s", err.Error())
	}
	log.Infof("'%s' Deployment Config rollout succeeded to latest", secretName)


	log.Infof("%s service created", component.Name)
	component.Status.Phase = v1alpha1.PhaseServiceCreation

	err = sdk.Update(component)
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func addSecretAsEnvFromSource(secretName string) []corev1.EnvFromSource {
	return []corev1.EnvFromSource{
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: secretName},
			},
		},
	}
}

func GetDeploymentConfig(namespace string, name string) (*deploymentconfig.DeploymentConfig, error) {
	dc := deploymentconfig.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind: "DeploymentConfig",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name: name,
		},
	}
	if err := sdk.Get(&dc); err != nil {
		return nil, err
	}
	return &dc, nil
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

			kubernetes.SetNamespaceAndOwnerReference(obj, component)
			result = append(result, obj)
		}
	} else {
		obj, err := k8sutil.RuntimeObjectFromUnstructured(r)
		if err != nil {
			return nil, err
		}

		kubernetes.SetNamespaceAndOwnerReference(obj, component)
		result = append(result, obj)
	}
	return result, nil
}
