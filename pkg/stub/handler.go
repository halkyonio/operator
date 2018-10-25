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
	"github.com/snowdrop/component-operator/pkg/stub/pipeline/link"
	"github.com/snowdrop/component-operator/pkg/stub/pipeline/servicecatalog"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	log "github.com/sirupsen/logrus"

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
		linkSteps: []pipeline.Step{
			link.NewLinkStep(),
		},
	}
}

type Handler struct {
	innerLoopSteps      []pipeline.Step
	serviceCatalogSteps []pipeline.Step
	linkSteps           []pipeline.Step
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {

	case *v1alpha1.Export:
		if o.Spec.Name != "" {
			log.Info("### Invoking export on ", o.Name)
		}

	case *v1alpha1.Component:
		// Deletion status
		deleted := event.Deleted
		if deleted {
			// the object `v1alpha1.Component` was deleted
			// handle the delete event here
			if o.Spec.Services != nil {
				for _, a := range h.serviceCatalogSteps {
					log.Infof("### Invoking pipeline 'service catalog', action 'delete' on %s", o.Name)
					if err := a.Handle(o, deleted); err != nil {
						return err
					}
				}
				break
			}
			return nil
		}
		// Check the DeploymentMode to install the component/runtime
		if o.Spec.Runtime != "" && o.Spec.DeploymentMode == "innerloop" {
			// log.Debug("DeploymentMode :", o.Spec.DeploymentMode)
			for _, a := range h.innerLoopSteps {
				if a.CanHandle(o) {
					log.Infof("### Invoking pipeline 'innerloop', action '%s' on %s", a.Name(), o.Name)
					if err := a.Handle(o, deleted); err != nil {
						return err
					}
				}
			}
			break
		}
		// Check if the component is a service
		if o.Spec.Services != nil {
			for _, a := range h.serviceCatalogSteps {
				if a.CanHandle(o) {
					log.Infof("### Invoking'service catalog', action '%s' on %s", a.Name(), o.Name)
					if err := a.Handle(o, deleted); err != nil {
						return err
					}
				}
			}
			break
		}
		// Check if the component is a Link
		if o.Spec.Link != nil {
			for _, a := range h.linkSteps {
				if a.CanHandle(o) {
					log.Infof("### Invoking'link', action '%s' on %s", a.Name(), o.Name)
					if err := a.Handle(o, deleted); err != nil {
						return err
					}
				}
			}
			break
		}
	}
	return nil
}
