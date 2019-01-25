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
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/snowdrop/component-operator/pkg/pipeline"
)

// Export step
func ExportStep() pipeline.Step {
	return &exportStep{}
}

type exportStep struct{}

func (exportStep) Name() string {
	return "export"
}

func (exportStep) CanHandle(component *v1alpha1.Component) bool {
	return true
}

func (exportStep) Handle(component *v1alpha1.Component, client *client.Client, namespace string, scheme *runtime.Scheme) error {
	return exportResources(component, *client, namespace, scheme)
}

func exportResources(component *v1alpha1.Component, c client.Client, namespace string, scheme *runtime.Scheme) error {
	component.ObjectMeta.Namespace = namespace
	//_ := new(runtime.Object)
	_ = client.ListOptions{
		Namespace:     component.ObjectMeta.Namespace,
		LabelSelector: getLabelsSelector(component.ObjectMeta.Labels),
	}
	return nil
}

func getLabelsSelector(mapLabels map[string]string) labels.Selector {
	return labels.SelectorFromSet(mapLabels)
}


