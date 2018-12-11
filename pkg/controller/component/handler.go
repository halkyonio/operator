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
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha1"
	"github.com/snowdrop/component-operator/pkg/pipeline"
	"github.com/snowdrop/component-operator/pkg/pipeline/innerloop"
	"github.com/snowdrop/component-operator/pkg/pipeline/link"
	"github.com/snowdrop/component-operator/pkg/pipeline/servicecatalog"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	. "github.com/snowdrop/component-operator/pkg/util/helper"
)

// Add creates a new AppService Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("component-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource AppService
	err = c.Watch(&source.Kind{Type: &v1alpha1.Component{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var (
	_                reconcile.Reconciler = &ReconcileComponent{}
	svcFinalizerName                      = "service.component.k8s.io"
)

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileComponent{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
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

type ReconcileComponent struct {
	client              client.Client
	scheme              *runtime.Scheme
	innerLoopSteps      []pipeline.Step
	serviceCatalogSteps []pipeline.Step
	linkSteps           []pipeline.Step
}

func (r *ReconcileComponent) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	operation := ""
	// Fetch the AppService instance
	component := &v1alpha1.Component{}
	err := r.client.Get(context.TODO(), request.NamespacedName, component)
	if err != nil {
		if errors.IsNotFound(err) {
			// Component has been deleted like also its dependencies
			operation = "deleted"
		}
		// Error reading the object
		log.Printf("Reconciling AppService %s/%s - operation %s\n", request.Namespace, request.Name, operation)
		return reconcile.Result{}, nil
	}

	// See finalizer doc for more info : https://book.kubebuilder.io/beyond_basics/using_finalizers.html
	// If DeletionTimeStamp is not equal zero, then the resource has been marked for deletion
	if !component.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is being deleted
		if ContainsString(component.ObjectMeta.Finalizers, svcFinalizerName) {
			// our finalizer is present, so lets handle our external dependency
			// Check if the component is a Service and then delete the ServiceInstance, ServiceBinding
			if component.Spec.Services != nil {
				removeServiceInstanceStep := servicecatalog.RemoveServiceInstanceStep()
				log.Infof("### Invoking'service catalog', action '%s' on %s", "delete", component.Name)
				if err := removeServiceInstanceStep.Handle(component, &r.client, request.Namespace); err != nil {
					log.Error(err)
				}
			}

			// remove our finalizer from the list and update it.
			component.ObjectMeta.Finalizers = RemoveString(component.ObjectMeta.Finalizers, svcFinalizerName)
			if err := r.client.Update(context.Background(), component); err != nil {
				return reconcile.Result{Requeue: true}, nil
			}
		}
	}

	// Component Custom Resource instance has been created
	operation = "created"
	log.Printf("Status : %s", component.Status.Phase)

	// Check if the component is a Service to be installed from the catalog
	if component.Spec.Services != nil {
		for _, a := range r.serviceCatalogSteps {
			if a.CanHandle(component) {
				log.Infof("### Invoking'service catalog', action '%s' on %s", a.Name(), component.Name)
				if err := a.Handle(component, &r.client, request.Namespace); err != nil {
					log.Error(err)
					return reconcile.Result{}, err
				}
			}
		}
	}

	// Check if Spec is not null and if the DeploymentMode strategy is equal to innerloop
	if component.Spec.Runtime != "" && component.Spec.DeploymentMode == "innerloop" {
		for _, a := range r.innerLoopSteps {
			if a.CanHandle(component) {
				log.Infof("### Invoking pipeline 'innerloop', action '%s' on %s", a.Name(), component.Name)
				if err := a.Handle(component, &r.client, request.Namespace); err != nil {
					log.Error(err)
					return reconcile.Result{}, err
				}
			}
		}
	}

	// Check if the component is a Link and that
	if component.Spec.Links != nil {
		for _, a := range r.linkSteps {
			if a.CanHandle(component) {
				log.Infof("### Invoking'link', action '%s' on %s", a.Name(), component.Name)
				if err := a.Handle(component, &r.client, request.Namespace); err != nil {
					log.Error(err)
					return reconcile.Result{}, err
				}
			}
		}
	}
	log.Printf("Status : %s", component.Status.Phase)
	log.Printf("Reconciling AppService %s/%s - operation %s\n", request.Namespace, request.Name, operation)
	return reconcile.Result{}, nil
}
