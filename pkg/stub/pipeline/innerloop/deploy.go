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
	"github.com/snowdrop/spring-boot-operator/pkg/apis/springboot/v1alpha1"
	"github.com/snowdrop/spring-boot-operator/pkg/stub/pipeline"
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

func (deployStep) CanHandle(springboot *v1alpha1.SpringBoot) bool {
	return true
}

func (deployStep) Handle(springboot *v1alpha1.SpringBoot) error {
	target := springboot.DeepCopy()
	return installDeployment(target)
}

func installDeployment(sb *v1alpha1.SpringBoot) error {
	// TODO
	return nil
}
