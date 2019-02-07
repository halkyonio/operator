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

package component

import (
	"context"
	"github.com/snowdrop/component-operator/pkg/pipeline/generic"
	"github.com/snowdrop/component-operator/pkg/pipeline/outerloop"
	"k8s.io/client-go/rest"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha1"
	"github.com/snowdrop/component-operator/pkg/pipeline"
	"github.com/snowdrop/component-operator/pkg/pipeline/innerloop"
	"github.com/snowdrop/component-operator/pkg/pipeline/link"
	"github.com/snowdrop/component-operator/pkg/pipeline/servicecatalog"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	. "github.com/snowdrop/component-operator/pkg/util/helper"
)

var (
	_                reconcile.Reconciler = &ReconcileComponent{}
	svcFinalizerName                      = "service.component.k8s.io"
	// Create a new instance of the logger. You can have any number of instances.
	log       = logrus.New()
	reconcileComponent = &ReconcileComponent{}
)

// Add creates a new Component Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return Create(mgr, NewReconciler(mgr))
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func Create(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("component-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Component
	err = c.Watch(&source.Kind{Type: &v1alpha1.Component{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// newReconciler returns a new reconcile.Reconciler
func NewReconciler(mgr manager.Manager) reconcile.Reconciler {
	rc := &ReconcileComponent{}
	rc.client = mgr.GetClient()
	rc.config = mgr.GetConfig()
	rc.scheme = mgr.GetScheme()
	rc.innerLoopSteps = []pipeline.Step{
		innerloop.NewInstallStep(),
	}
	rc.outerLoopSteps = []pipeline.Step{
		outerloop.NewInstallStep(),
		outerloop.NewCloneDeploymentStep(),
		generic.NewUpdateServiceSelectorStep(),
	}
	rc.serviceCatalogSteps = []pipeline.Step{
		servicecatalog.NewServiceInstanceStep(),
	}
	rc.linkSteps = []pipeline.Step{
		link.NewLinkStep(),
	}
	reconcileComponent = rc
	return rc
}
type ReconcileComponent struct {
	client              client.Client
	config              *rest.Config
	scheme              *runtime.Scheme
	innerLoopSteps      []pipeline.Step
	outerLoopSteps      []pipeline.Step
	serviceCatalogSteps []pipeline.Step
	linkSteps           []pipeline.Step
}

func (r *ReconcileComponent) Reconcile(request reconcile.Request) (reconcile.Result, error) {

	operation := ""

	// Fetch the Component created, deleted or updated
	component := &v1alpha1.Component{}
	err := r.client.Get(context.TODO(), request.NamespacedName, component)

	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// Finalizer has been removed and component deleted. So we can exit
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	log.Info("----------------------------------------------------")
	log.Infof("***** Reconciling Component %s, namespace %s", request.Name, request.Namespace)
	log.Infof("** Status of the component : %s", component.Status.Phase)
	log.Infof("** Creation time : %s", component.ObjectMeta.CreationTimestamp)
	log.Infof("** Resource version : %s", component.ObjectMeta.ResourceVersion)
	log.Infof("** Generation version : %s", strconv.FormatInt(component.ObjectMeta.Generation, 10))
	log.Infof("** Deletion time : %s", component.ObjectMeta.DeletionTimestamp)
	log.Info("----------------------------------------------------")

	// Assign the generated ResourceVersion to the resource status
	if component.Status.RevNumber == "" {
		component.Status.RevNumber = component.ObjectMeta.ResourceVersion
	}

	// See finalizer doc for more info : https://book.kubebuilder.io/beyond_basics/using_finalizers.html
	// If DeletionTimeStamp is not equal zero, then the resource has been marked for deletion
	if !component.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is being deleted
		if ContainsString(component.ObjectMeta.Finalizers, svcFinalizerName) {
			// Component has been deleted like also its dependencies
			operation = "DELETION"

			// our finalizer is present, so lets handle our external dependency
			// Check if the component is a Service and then delete the ServiceInstance, ServiceBinding
			if component.Spec.Services != nil {
				removeServiceInstanceStep := servicecatalog.RemoveServiceInstanceStep()
				log.Infof("## Invoking'service catalog', action '%s' on %s", "delete", component.Name)
				//log.Infof("## Invoking'service catalog', action '%s' on %s", "delete", component.Name)
				if err := removeServiceInstanceStep.Handle(component, r.config, &r.client, request.Namespace, r.scheme); err != nil {
					log.Errorf("Removing Service Instance, binding failed %s", err)
				}
			}

			// remove our finalizer from the list and update it.
			component.ObjectMeta.Finalizers = RemoveString(component.ObjectMeta.Finalizers, svcFinalizerName)
			if err := r.client.Update(context.Background(), component); err != nil {
				return reconcile.Result{Requeue: true}, nil
			}
		}
		log.Infof("***** Reconciled Component %s, namespace %s", request.Name, request.Namespace)
		log.Infof("***** Operation performed : %s", operation)
		return reconcile.Result{}, nil
	}

	// We only call the pipeline when the component has been created
	// and if the Status Revision Number is the same
	if component.Status.RevNumber == component.ObjectMeta.ResourceVersion {

		// Component Custom Resource instance has been created
		operation = "CREATION"

		// Check if Spec is not null and if the DeploymentMode strategy is equal to innerloop
		if component.Spec.Runtime != "" && component.Spec.DeploymentMode == "innerloop" {
			for _, a := range r.innerLoopSteps {
				if a.CanHandle(component) {
					log.Infof("## Invoking pipeline 'innerloop', action '%s' on %s", a.Name(), component.Name)
					if err := a.Handle(component, r.config, &r.client, request.Namespace, r.scheme); err != nil {
						log.Error("Innerloop creation failed", err)
						return reconcile.Result{}, err
					}
				}
			}
		}

		// Check if the component is a Service to be installed from the catalog
		if component.Spec.Services != nil {
			for _, a := range r.serviceCatalogSteps {
				if a.CanHandle(component) {
					log.Infof("## Invoking pipeline 'service catalog', action '%s' on %s", a.Name(), component.Name)
					if err := a.Handle(component, r.config, &r.client, request.Namespace, r.scheme); err != nil {
						log.Error("Service instance, binding creation failed", err)
						return reconcile.Result{}, err
					}
				}
			}
		}

		// Check if the component is a Link and that
		if component.Spec.Links != nil {
			for _, a := range r.linkSteps {
				if a.CanHandle(component) {
					log.Infof("## Invoking pipeline 'link', action '%s' on %s", a.Name(), component.Name)
					if err := a.Handle(component, r.config, &r.client, request.Namespace, r.scheme); err != nil {
						log.Error("Linking components failed", err)
						return reconcile.Result{}, err
					}
				}
			}
		}
	} else {
		operation = "UPDATE"

		if component.Spec.DeploymentMode == "outerloop" {
			for _, a := range r.outerLoopSteps {
				if a.CanHandle(component) {
					log.Infof("## Invoking pipeline 'outerloop', action '%s' on %s", a.Name(), component.Name)
					if err := a.Handle(component, r.config, &r.client, request.Namespace, r.scheme); err != nil {
						log.Error("Outerloop creation failed", err)
						return reconcile.Result{}, err
					}
				}
			}
		} else {
			log.Info("No pipeline invoked")
			log.Info("------------------------------------------------------")
		}
	}

	log.Infof("***** Reconciled Component %s, namespace %s", request.Name, request.Namespace)
	log.Infof("***** Operation performed : %s", operation)
	return reconcile.Result{}, nil
}
