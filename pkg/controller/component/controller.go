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
	"fmt"
	"github.com/go-logr/logr"
	"golang.org/x/net/context"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"
	"strconv"

	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	// . "github.com/snowdrop/component-operator/pkg/util/helper"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const (
	svcFinalizerName  = "service.devexp.runtime.redhat.com"
	controllerName    = "component-controller"
	deletionOperation = "DELETION"
	creationOperation = "CREATION"
	updateOperation   = "UPDATE"

	DEPLOYMENT       = "Deployment"
	DEPLOYMENTCONFIG = "DeploymentConfig"
	SERVICE          = "Capability"
	ROUTE            = "Route"
	INGRESS          = "Ingress"
	IMAGESTREAM      = "ImageStream"
	IMAGESTREAMLIST  = "ImageStreamList"
	BUILDCONFIG      = "BuildConfig"
	PERSISTENTVOLUMECLAIM      = "BuildConfig"
)


var log = logf.Log.WithName("component.controller")

// New creates a new Component Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func New(mgr manager.Manager) error {
	return Add(mgr, NewReconciler(mgr))
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func Add(mgr manager.Manager, r reconcile.Reconciler) error {
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
	//BuildConfig
	if err := watchBuildConfig(c); err != nil {
		return err
	}

	//Deployment
	if err := watchDeployment(c); err != nil {
		return err
	}

	//DeploymentConfig
	if err := watchDeploymentConfig(c); err != nil {
		return err
	}

	//Pod
	if err := watchPod(c); err != nil {
		return err
	}

	//Capability
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
	}
}

type ReconcileComponent struct {
	client              client.Client
	config              *rest.Config
	scheme              *runtime.Scheme
	reqLogger           logr.Logger
}

//buildFactory will return the resource according to the kind defined
func (r *ReconcileComponent) buildFactory(instance *v1alpha2.Component, kind string) (runtime.Object, error) {
	r.reqLogger.Info("Check "+kind, "into the namespace", instance.Namespace)
	switch kind {
	case DEPLOYMENTCONFIG:
		return r.buildDeploymentConfig(instance), nil
	case DEPLOYMENT:
		return r.buildDeployment(instance), nil
	case SERVICE:
		return r.buildService(instance), nil
	case ROUTE:
		return r.buildRoute(instance), nil
	case INGRESS:
		return r.buildRoute(instance), nil
	case PERSISTENTVOLUMECLAIM:
		return r.buildPVC(instance), nil
	default:
		msg := "Failed to recognize type of object" + kind + " into the Namespace " + instance.Namespace
		panic(msg)
	}
}

//Create the factory object and requeue
func (r *ReconcileComponent) create(instance *v1alpha2.Component, kind string, err error) (reconcile.Result, error) {
	obj, errBuildObject := r.buildFactory(instance, kind)
	if errBuildObject != nil {
		return reconcile.Result{}, errBuildObject
	}
	if errors.IsNotFound(err) {
		r.reqLogger.Info("Creating a new ", "kind", kind, "Namespace", instance.Namespace)
		err = r.client.Create(context.TODO(), obj)
		if err != nil {
			r.reqLogger.Error(err, "Failed to create new ", "kind", kind, "Namespace", instance.Namespace)
			return reconcile.Result{}, err
		}
		r.reqLogger.Info("Created successfully - return and create", "kind", kind, "Namespace", instance.Namespace)
		return reconcile.Result{Requeue: true}, nil
	}
	r.reqLogger.Error(err, "Failed to get", "kind", kind, "Namespace", instance.Namespace)
	return reconcile.Result{}, err

}

func (r *ReconcileComponent) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	r.reqLogger = log.WithValues("Namespace",request.Namespace)

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

	r.reqLogger.Info("-----------------------")
	r.reqLogger.Info("Reconciling Component  ")
	r.reqLogger.Info("Status of the component","Status phase", component.Status.Phase)
	r.reqLogger.Info("Creation time          ","Creation time", component.ObjectMeta.CreationTimestamp)
	r.reqLogger.Info("Resource version       ","Resource version", component.ObjectMeta.ResourceVersion)
	r.reqLogger.Info("Generation version     ","Generation version", strconv.FormatInt(component.ObjectMeta.Generation, 10))
	// r.reqLogger.Info("Deletion time          ","Deletion time", component.ObjectMeta.DeletionTimestamp)

	// Add the Status Component Creation when we process the first time the Component CR
	// as we will start to create different resources
	if component.Generation == 1 && component.Status.Phase == "" {
		if err := r.updateComponentStatus(component, v1alpha2.PhaseComponentCreation, request); err != nil {
			r.reqLogger.Info("Status update failed !")
			return reconcile.Result{}, err
		}
	}

	switch m := component.Spec.DeploymentMode; m {
	case "innerloop":
		if err := r.installInnerLoop(component, request.Namespace); err != nil {
			r.reqLogger.Error(err, "Innerloop creation failed")
			return reconcile.Result{}, err
		}
	case "outerloop":
		if err := r.installIOuterLoop(component, request.Namespace); err != nil {
			r.reqLogger.Error(err, "Outerloop creation failed")
			return reconcile.Result{}, err
		}
	default:
		if err := r.installInnerLoop(component, request.Namespace); err != nil {
			r.reqLogger.Error(err, "Innerloop creation failed")
			return reconcile.Result{}, err
		}
	}

	//Update Pod Status
	podStatus, err := r.updatePodStatus(component,request)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Update status of the Component
	if err := r.updateStatus(podStatus, component, request); err != nil {
		return reconcile.Result{}, err
	}

	r.reqLogger.Info(fmt.Sprintf("Reconciled : %s",component.Name))
	return reconcile.Result{}, nil
}
