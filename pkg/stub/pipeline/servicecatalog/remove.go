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
	servicecatalog "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	log "github.com/sirupsen/logrus"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha1"
	"github.com/snowdrop/component-operator/pkg/stub/pipeline"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RemoveServiceInstanceStep creates a step that handles the creation of the Service from the catalog
func RemoveServiceInstanceStep() pipeline.Step {
	return &removeServiceInstanceStep{}
}

type removeServiceInstanceStep struct {
}

func (removeServiceInstanceStep) Name() string {
	return "remove service"
}

func (removeServiceInstanceStep) CanHandle(component *v1alpha1.Component) bool {
	return component.Status.Phase == "CreatingService"
}

func (removeServiceInstanceStep) Handle(component *v1alpha1.Component, deleted bool) error {
	target := component.DeepCopy()
	if deleted {
		return deleteService(target)
	}
	return nil
}

func deleteService(component *v1alpha1.Component) error {
	selector := getComponentSelector()
	for _, s := range component.Spec.Services {
		// Let's retrieve the ServiceBindings to delete them first
		list, err := listServiceBindings(component, selector)
		if err != nil {
			return err
		}
		// Delete ServiceBinding(s) linked to the ServiceInstance
		for _, sb := range list.Items {
			if sb.Name == s.Name {
				err := sdk.Delete(&sb)
				if err != nil {
					return err
				}
				log.Infof("#### Deleted serviceBinding '%s' for the service '%s'", sb.Name, s.Name)
			}
		}

		// Retrieve ServiceInstances
		list = new(servicecatalog.ServiceInstanceList)
		list.TypeMeta = metav1.TypeMeta{
			Kind:       "ServiceInstance",
			APIVersion: "servicecatalog.k8s.io/v1beta1",
		}
		err = sdk.List(component.ObjectMeta.Namespace, list)
		if err != nil {
			return err
		}

		// Delete ServiceInstance(s)
		for _, si := range list.Items {
			err := sdk.Delete(&si)
			if err != nil {
				return err
			}
			log.Infof("#### Deleted serviceInstance '%s' for the service '%s'", si.Name, s.Name)
		}
	}
	return nil
}