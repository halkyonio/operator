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
	"github.com/go-logr/logr"
	"github.com/snowdrop/component-operator/pkg/pipeline/generic"
	"github.com/snowdrop/component-operator/pkg/pipeline/outerloop"
	"golang.org/x/net/context"
	"k8s.io/client-go/rest"
	"strconv"

	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"github.com/snowdrop/component-operator/pkg/pipeline"
	"github.com/snowdrop/component-operator/pkg/pipeline/link"
	"github.com/snowdrop/component-operator/pkg/pipeline/servicecatalog"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	. "github.com/snowdrop/component-operator/pkg/util/helper"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const (
	svcFinalizerName  = "service.component.k8s.io"
	controllerName    = "component-controller"
	deletionOperation = "DELETION"
	creationOperation = "CREATION"
	updateOperation   = "UPDATE"
)

var log = logf.Log.WithName("controller_component")

// New creates a new Component Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func New(mgr manager.Manager) error {
	return create(mgr, NewReconciler(mgr))
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func create(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New(controllerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Component
	err = c.Watch(&source.Kind{Type: &v1alpha2.Component{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	/** Watch for changes of child/secondary resources **/
	//Deployment
	if err := watchDeployment(c); err != nil {
		return err
	}

	//Service
	if err := watchService(c); err != nil {
		return err
	}

	//Route
	if err:= watchRoute(c); err != nil {
		return err
	}

	return nil
}

// newReconciler returns a new reconcile.Reconciler
func NewReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileComponent{
		client: mgr.GetClient(),
		config: mgr.GetConfig(),
		scheme: mgr.GetScheme(),
		reqLogger: log,
		outerLoopSteps: []pipeline.Step{
			outerloop.NewInstallStep(),
			outerloop.NewCloneDeploymentStep(),
			generic.NewUpdateServiceSelectorStep(),
		},
		serviceCatalogSteps: []pipeline.Step{
			servicecatalog.NewServiceInstanceStep(),
		},
		linkSteps: []pipeline.Step{
			link.NewLinkStep(),
		},
	}
}

type ReconcileComponent struct {
	client              client.Client
	config              *rest.Config
	scheme              *runtime.Scheme
	reqLogger           logr.Logger
	outerLoopSteps      []pipeline.Step
	serviceCatalogSteps []pipeline.Step
	linkSteps           []pipeline.Step
}

func (r *ReconcileComponent) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)

	operation := ""

	// Fetch the Component created, deleted or updated
	component := &v1alpha2.Component{}
	err := r.client.Get(context.TODO(), request.NamespacedName, component)
	if err != nil {
		return r.fetch(err)
	}

	//TODO : add Check/Validate spec content
	if !r.hasMandatorySpecs(component) {
		return reconcile.Result{Requeue: true}, nil
	}

	reqLogger.Info("----------------------------------------------------")
	reqLogger.Info("***** Reconciling Component ******")
	reqLogger.Info("** Status of the component", "Status phase", component.Status.Phase)
	reqLogger.Info("** Creation time","Creation time", component.ObjectMeta.CreationTimestamp)
	reqLogger.Info("** Resource version", "Resource version", component.ObjectMeta.ResourceVersion)
	reqLogger.Info("** Generation version", "Generation version", strconv.FormatInt(component.ObjectMeta.Generation, 10))
	reqLogger.Info("** Deletion time", "Deletion time", component.ObjectMeta.DeletionTimestamp)
	reqLogger.Info("----------------------------------------------------")

	// Check if the child resources needed are created according to the mode
	// Check if Spec is not null and if the DeploymentMode strategy is equal to Dev Mode (aka innerloop)
	if component.Spec.Runtime != "" && component.Spec.DeploymentMode == "innerloop" {
		if err := r.installInnerLoop(component, request.Namespace); err != nil {
			reqLogger.Error(err, "Innerloop creation failed")
			return reconcile.Result{}, err
		}
	}

	// TODO: Align the logic hereafter as we did for the inner loop
	if component.Spec.DeploymentMode == "outerloop" {
		for _, a := range r.outerLoopSteps {
			if a.CanHandle(component) {
				reqLogger.Info("## Invoking pipeline 'outerloop'","Action", a.Name(), "Component name", component.Name)
				if err := a.Handle(component, r.config, &r.client, request.Namespace, r.scheme); err != nil {
					reqLogger.Error(err, "Outerloop creation failed")
					return reconcile.Result{}, err
				}
			}
		}
	}

	// Check if the component is a Service to be installed from the catalog
	if component.Spec.Services != nil {
		for _, a := range r.serviceCatalogSteps {
			if a.CanHandle(component) {
				reqLogger.Info("## Invoking pipeline 'service catalog', action '%s' on %s", a.Name(), component.Name)
				if err := a.Handle(component, r.config, &r.client, request.Namespace, r.scheme); err != nil {
					reqLogger.Error(err,"Service instance, binding creation failed")
					return reconcile.Result{}, err
				}
			}
		}
	}

	// Check if the component is a Link and that
	if component.Spec.Links != nil {
		for _, a := range r.linkSteps {
			if a.CanHandle(component) {
				reqLogger.Info("## Invoking pipeline 'link', action '%s' on %s", a.Name(), component.Name)
				if err := a.Handle(component, r.config, &r.client, request.Namespace, r.scheme); err != nil {
					reqLogger.Error(err,"Linking components failed")
					return reconcile.Result{}, err
				}
			}
		}
	}

	// See finalizer doc for more info : https://book.kubebuilder.io/beyond_basics/using_finalizers.html
	// If DeletionTimeStamp is not equal zero, then the resource has been marked for deletion
	if !component.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is being deleted
		if ContainsString(component.ObjectMeta.Finalizers, svcFinalizerName) {
			// Component has been deleted like also its dependencies
			operation = deletionOperation

			// our finalizer is present, so lets handle our external dependency
			// Check if the component is a Service and then delete the ServiceInstance, ServiceBinding
			// TODO: Move this code under the ServiceController !!
			if component.Spec.Services != nil {
				removeServiceInstanceStep := servicecatalog.RemoveServiceInstanceStep()
				reqLogger.Info("## Invoking'service catalog', action '%s' on %s", "delete", component.Name)
				//log.Infof("## Invoking'service catalog', action '%s' on %s", "delete", component.Name)
				if err := removeServiceInstanceStep.Handle(component, r.config, &r.client, request.Namespace, r.scheme); err != nil {
					reqLogger.Error(err, "Removing Service Instance, binding failed")
				}
			}

			// remove our finalizer from the list and update it.
			component.ObjectMeta.Finalizers = RemoveString(component.ObjectMeta.Finalizers, svcFinalizerName)
			if err := r.client.Update(context.Background(), component); err != nil {
				return reconcile.Result{Requeue: true}, nil
			}
		}
		reqLogger.Info("***** Reconciled Component %s, namespace %s", request.Name, request.Namespace)
		reqLogger.Info("***** Operation performed : %s", operation)
		return reconcile.Result{}, nil
	}

	reqLogger.Info("***** Reconciled Component *****")
	reqLogger.Info("***** Action ", "Type", operation)
	return reconcile.Result{}, nil
}
