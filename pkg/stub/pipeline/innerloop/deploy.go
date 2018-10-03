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
	"github.com/ghodss/yaml"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha1"
	"github.com/snowdrop/component-operator/pkg/stub/pipeline"
	"github.com/snowdrop/component-operator/pkg/types"
	"github.com/snowdrop/component-operator/pkg/util/template"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewDeployStep creates a step that handles the creation of the DeploymentConfig
func NewDeployStep() pipeline.Step {
	return &deployStep{}
}

type deployStep struct {
}

func (deployStep) Name() string {
	return "deploy"
}

func (deployStep) CanHandle(component *v1alpha1.Component) bool {
	return true
}

func (deployStep) Handle(component *v1alpha1.Component) error {
	target := component.DeepCopy()
	return installDeploymentConfig(target)
}

func installDeploymentConfig(component *v1alpha1.Component) error {
	// Create Route
	//route := newComponentRoute(component.Name)
	route := newComponentRouteFromTemplate(component.Name)
	err := sdk.Create(route)
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func newComponentRouteFromTemplate(name string) *routev1.Route {
	// Parse Route Template
	var b = template.ParseTemplate(template.GetTemplateFullName("route"), types.Application{Name: name})

	// Create Route struct using the generated Route string
	route := routev1.Route{}
	err := yaml.Unmarshal(b.Bytes(), &route)
	if err != nil {
		panic(err)
	}
	route.Namespace = "component-operator"
	return &route
}

func newComponentRoute(name string) *routev1.Route {
	return &routev1.Route{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Route",
			APIVersion: "route.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Labels:    getLabels(name),
			Namespace: "component-operator",
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: name,
			},
		},
	}
}

func getLabels(component string) map[string]string {
	labels := map[string]string{
		"Component": component,
	}
	return labels
}
