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
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha1"
	"github.com/snowdrop/component-operator/pkg/stub/pipeline"
)

// NewPVCStep creates a step that handles the creation of the DeploymentConfig
func NewPVCStep() pipeline.Step {
	return &pvcStep{}
}

type pvcStep struct {
}

func (pvcStep) Name() string {
	return "pvc"
}

func (pvcStep) CanHandle(component *v1alpha1.Component) bool {
	return true
}

func (pvcStep) Handle(component *v1alpha1.Component) error {
	return nil
}
