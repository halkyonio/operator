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

package generic

import (
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/snowdrop/spring-boot-operator/pkg/apis/springboot/v1alpha1"
	"github.com/snowdrop/spring-boot-operator/pkg/stub/pipeline"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// NewServiceStep creates a step that handles the creation of the DeploymentConfig
func NewServiceStep() pipeline.Step {
	return &serviceStep{}
}

type serviceStep struct {
}

func (serviceStep) Name() string {
	return "service"
}

func (serviceStep) CanHandle(integration *v1alpha1.SpringBoot) bool {
	return true
}

func (serviceStep) Handle(integration *v1alpha1.SpringBoot) error {
	// TODO : Refactor to create a real service
	serviceName := "toto"
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   serviceName,
			Labels: map[string]string{"name": serviceName},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:     8080,
					Protocol: corev1.ProtocolTCP,
					TargetPort: intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "8080",
					},
					Name: serviceName,
				},
			},
			Selector: map[string]string{"name": serviceName},
		},
	}
	err := sdk.Get(service)
	if err != nil {
		return errors.Wrap(err, "could not get service for integration ")
	}
	log.Info("Get service")

	return nil
}
