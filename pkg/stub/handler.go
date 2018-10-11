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
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha1"
	"github.com/snowdrop/component-operator/pkg/stub/pipeline"
	"github.com/snowdrop/component-operator/pkg/stub/pipeline/innerloop"
	"github.com/snowdrop/component-operator/pkg/stub/pipeline/servicecatalog"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"

	// import openshift package to register the OpenShift schemes (route, image, ...)
	_ "github.com/snowdrop/component-operator/pkg/util/openshift"
)

func NewHandler() sdk.Handler {
	return &Handler{
		innerLoopSteps: []pipeline.Step{
			innerloop.NewInstallStep(),
		},
		serviceCatalogSteps: []pipeline.Step{
			servicecatalog.NewServiceInstanceStep(),
		},
	}
}

type Handler struct {
	innerLoopSteps      []pipeline.Step
	serviceCatalogSteps []pipeline.Step
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1alpha1.Component:
		// Check the DeploymentMode to install the component/runtime
		if o.Spec.DeploymentMode == "innerloop" && o.Spec.Runtime != "" {
			logrus.Debug("DeploymentMode :", o.Spec.DeploymentMode)
			for _, a := range h.innerLoopSteps {
				if a.CanHandle(o) {
					logrus.Debug("Invoking action ", a.Name(), " on Spring Boot ", o.Name)
					if err := a.Handle(o); err != nil {
						return err
					}
				}
			}
		}
		// Check if the component is a service
		if o.Spec.Services != nil {
			for _, a := range h.serviceCatalogSteps {
				if a.CanHandle(o) {
					logrus.Debug("Invoking action ", a.Name(), " on Spring Boot ", o.Name)
					if err := a.Handle(o); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
