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

package innerloop

import (
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha1"
	"github.com/snowdrop/component-operator/pkg/stub/pipeline"
	"github.com/snowdrop/component-operator/pkg/util/kubernetes"
	util "github.com/snowdrop/component-operator/pkg/util/template"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"text/template"
)

var (
	namespace = "my-spring-app"
)

// NewInstallStep creates a step that handles the creation of the DeploymentConfig
func NewInstallStep() pipeline.Step {
	return &installStep{}
}

type installStep struct {
}

func (installStep) Name() string {
	return "deploy"
}

func (installStep) CanHandle(component *v1alpha1.Component) bool {
	return component.Status.Phase == ""
}

func (installStep) Handle(component *v1alpha1.Component) error {
	target := component.DeepCopy()
	return installInnerLoop(target)
}

func installInnerLoop(component *v1alpha1.Component) error {
	// TODO Add a key to get the templates associated to the innerloop, ....
	for _, tmpl := range util.Templates {
		res, err := newResourceFromTemplate(tmpl, component, namespace)
		if err != nil {
			return err
		}
		err = sdk.Create(res)
		if err != nil && !k8serrors.IsAlreadyExists(err) {
			return err
		}
	}
	return nil
}

func newResourceFromTemplate(template template.Template, component *v1alpha1.Component, namespace string) (runtime.Object, error) {
	var b = util.Parse(template, component)

	obj, err := kubernetes.PopulateKubernetesObjectFromYaml(b.String())
	if err != nil {
		return nil, err
	}

	// Define the namespace
	if metaObject, ok := obj.(metav1.Object); ok {
		metaObject.SetNamespace(namespace)
	}
	return obj, nil
}

func getLabels(component string) map[string]string {
	labels := map[string]string{
		"Component": component,
	}
	return labels
}
