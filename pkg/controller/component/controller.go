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
	routev1 "github.com/openshift/api/route/v1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"golang.org/x/net/context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	controllerName = "component-controller"
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
	//Deployment
	if err := watchDeployment(c); err != nil {
		return err
	}

	//Pod
	if err := watchPod(c); err != nil {
		return err
	}

	//Service
	if err := watchService(c); err != nil {
		return err
	}

	return nil
}

// newReconciler returns a new reconcile.Reconciler
func NewReconciler(mgr manager.Manager) reconcile.Reconciler {
	// todo: make this configurable
	images := make(map[string]imageInfo, 7)
	defaultEnvVar := make(map[string]string, 7)
	defaultEnvVar["JAVA_APP_DIR"] = "/deployment"
	defaultEnvVar["JAVA_DEBUG"] = "false"
	defaultEnvVar["JAVA_DEBUG_PORT"] = "5005"
	defaultEnvVar["JAVA_APP_JAR"] = "app.jar"
	images["spring-boot"] = imageInfo{
		registryRef: "quay.io/snowdrop/spring-boot-s2i",
		defaultEnv:  defaultEnvVar,
	}
	images["vert.x"] = imageInfo{
		registryRef: "quay.io/snowdrop/spring-boot-s2i",
		defaultEnv:  defaultEnvVar,
	}
	images["thorntail"] = imageInfo{
		registryRef: "quay.io/snowdrop/spring-boot-s2i",
		defaultEnv:  defaultEnvVar,
	}
	// References images
	images["openjdk8"] = imageInfo{registryRef: "registry.access.redhat.com/redhat-openjdk-18/openjdk18-openshift"}
	images["nodejs"] = imageInfo{registryRef: "nodeshift/centos7-s2i-nodejs"}
	images[supervisorImageId] = imageInfo{registryRef: "quay.io/snowdrop/supervisord"}

	supervisor := v1alpha2.Component{
		ObjectMeta: v1.ObjectMeta{
			Name: supervisorContainerName,
		},
		Spec: v1alpha2.ComponentSpec{
			Runtime: supervisorImageId,
			Version: latestVersionTag,
			Envs: []v1alpha2.Env{
				{
					Name: "CMDS",
					Value: "run-java:/usr/local/s2i/run;run-node:/usr/libexec/s2i/run;compile-java:/usr/local/s2i" +
						"/assemble;build:/deployments/buildapp",
				},
			},
		},
	}

	r := &ReconcileComponent{
		client:             mgr.GetClient(),
		config:             mgr.GetConfig(),
		scheme:             mgr.GetScheme(),
		reqLogger:          log,
		runtimeImages:      images,
		supervisor:         &supervisor,
		dependentResources: make(map[string]dependentResource, 11),
	}

	r.addDependentResource(&corev1.PersistentVolumeClaim{}, r.buildPVC, func(c *v1alpha2.Component) string {
		specified := c.Spec.Storage.Name
		if len(specified) > 0 {
			return specified
		}
		return "m2-data-" + c.Name
	})
	r.addDependentResource(&appsv1.Deployment{}, r.buildDeployment, defaultNamer)
	r.addDependentResource(&corev1.Service{}, r.buildService, defaultNamer)
	r.addDependentResource(&corev1.ServiceAccount{}, r.buildServiceAccount, func(c *v1alpha2.Component) string {
		return serviceAccountName
	})
	r.addDependentResource(&routev1.Route{}, r.buildRoute, defaultNamer)
	r.addDependentResource(&v1beta1.Ingress{}, r.buildIngress, defaultNamer)
	taskNamer := func(c *v1alpha2.Component) string {
		return taskS2iBuildahPusName
	}
	r.addDependentResource(&v1alpha1.Task{}, r.buildTaskS2iBuildahPush, taskNamer)
	r.addDependentResource(&v1alpha1.TaskRun{}, r.buildTaskRunS2iBuildahPush, taskNamer)

	return r
}

func (r *ReconcileComponent) addDependentResource(res runtime.Object, buildFn builder, nameFn namer) {
	key, kind := getKeyAndKindFor(res)
	r.dependentResources[key] = dependentResource{
		build:     buildFn,
		name:      nameFn,
		prototype: res,
		fetch:     r.genericFetcher,
		kind:      kind,
	}
}

type imageInfo struct {
	registryRef string
	defaultEnv  map[string]string
}

var defaultNamer namer = func(component *v1alpha2.Component) string {
	return component.Name
}

type namer func(*v1alpha2.Component) string
type builder func(dependentResource, *v1alpha2.Component) (runtime.Object, error)
type fetcher func(dependentResource, *v1alpha2.Component) (runtime.Object, error)

func (r *ReconcileComponent) genericFetcher(res dependentResource, c *v1alpha2.Component) (runtime.Object, error) {
	into := res.prototype.DeepCopyObject()
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: res.name(c), Namespace: c.Namespace}, into); err != nil {
		r.reqLogger.Info(res.kind + " doesn't exist")
		return nil, err
	}
	return into, nil
}

type dependentResource struct {
	name      namer
	build     builder
	fetch     fetcher
	prototype runtime.Object
	kind      string
}

type ReconcileComponent struct {
	client             client.Client
	config             *rest.Config
	scheme             *runtime.Scheme
	reqLogger          logr.Logger
	runtimeImages      map[string]imageInfo
	supervisor         *v1alpha2.Component
	onOpenShift        *bool
	dependentResources map[string]dependentResource
}

//Create the factory object
func (r *ReconcileComponent) createIfNeeded(instance *v1alpha2.Component, resourceType runtime.Object) (bool, error) {
	key, kind := getKeyAndKindFor(resourceType)
	resource, ok := r.dependentResources[key]
	if !ok {
		return false, fmt.Errorf("unknown dependent type %s", kind)
	}

	if _, err := resource.fetch(resource, instance); err != nil {
		// create the object
		obj, errBuildObject := resource.build(resource, instance)
		if errBuildObject != nil {
			return false, errBuildObject
		}
		if errors.IsNotFound(err) {
			err = r.client.Create(context.TODO(), obj)
			if err != nil {
				r.reqLogger.Error(err, "Failed to create new ", "kind", kind)
				return false, err
			}
			r.reqLogger.Info("Created successfully", "kind", kind)
			return true, nil
		}
		r.reqLogger.Error(err, "Failed to get", "kind", kind)
		return false, err
	} else {
		return false, nil
	}
}

func getKeyAndKindFor(resourceType runtime.Object) (key string, kind string) {
	t := reflect.TypeOf(resourceType)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	pkg := t.PkgPath()
	kind = t.Name()
	key = pkg + "/" + kind
	return
}

func (r *ReconcileComponent) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	r.reqLogger = log.WithValues("namespace", request.Namespace)

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

	r.reqLogger.Info("==> Reconciling Component",
		"name", component.Name,
		"status", component.Status.Phase,
		"created", component.ObjectMeta.CreationTimestamp)

	// Add the Status Component Creation when we process the first time the Component CR
	// as we will start to create different resources
	if component.Generation == 1 && component.Status.Phase == "" {
		if err := r.updateStatus(component, v1alpha2.ComponentPending); err != nil {
			r.reqLogger.Info("Status update failed !")
			return reconcile.Result{}, err
		}
	}

	installFn := r.installDevMode
	r.setInitialStatus(component, v1alpha2.ComponentPending)
	if v1alpha2.BuildDeploymentMode == component.Spec.DeploymentMode {
		r.setInitialStatus(component, v1alpha2.ComponentBuilding)
		installFn = r.installBuildMode
	}

	result, err := r.installAndUpdateStatus(component, request, installFn)
	r.reqLogger.Info("<== Reconciled Component", "name", component.Name)
	return result, err
}

type installFnType func(component *v1alpha2.Component, namespace string) (bool, error)

func (r *ReconcileComponent) installAndUpdateStatus(component *v1alpha2.Component, request reconcile.Request, install installFnType) (reconcile.Result, error) {
	changed, err := install(component, request.Namespace)
	if err != nil {
		r.reqLogger.Error(err, fmt.Sprintf("failed to install %s mode", component.Spec.DeploymentMode))
		r.setErrorStatus(component, err)
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: changed}, r.updateStatus(component, v1alpha2.ComponentReady)
}

func (r *ReconcileComponent) setInitialStatus(component *v1alpha2.Component, phase v1alpha2.ComponentPhase) error {
	if component.Generation == 1 && component.Status.Phase == "" {
		if err := r.updateStatus(component, phase); err != nil {
			r.reqLogger.Info("Status update failed !")
			return err
		}
	}
	return nil
}
