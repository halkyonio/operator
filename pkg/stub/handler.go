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

package stub

import (
	"context"
	"github.com/snowdrop/spring-boot-operator/pkg/apis/springboot/v1alpha1"
	"github.com/snowdrop/spring-boot-operator/pkg/stub/pipeline/installation"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
)

func NewHandler() sdk.Handler {
	return &Handler{
		installationSteps: []installation.Step{
			//installation.NewDeployStep(),
			installation.NewServiceStep(),
			//installation.NewRouteStep(),
		},
	}
}

type Handler struct {
	installationSteps []installation.Step
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1alpha1.SpringBoot:
		for _, a := range h.installationSteps {
			if a.CanHandle(o) {
				logrus.Debug("Invoking action ", a.Name(), " on Spring Boot ", o.Name)
				if err := a.Handle(o); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
