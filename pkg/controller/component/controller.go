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
	routev1 "github.com/openshift/api/route/v1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	controller2 "github.com/snowdrop/component-operator/pkg/controller"
	"golang.org/x/net/context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// newReconciler returns a new reconcile.Reconciler
func NewComponentReconciler(mgr manager.Manager) *ReconcileComponent {
	// todo: make this configurable
	images := make(map[string]imageInfo, 7)
	defaultEnvVar := make(map[string]string, 7)
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
		runtimeImages:      images,
		supervisor:         &supervisor,
		dependentResources: make(map[string]dependentResource, 11),
	}
	r.ReconcilerHelper = controller2.NewHelper(r.PrimaryResourceType(), mgr)

	r.initDependentResources()

	return r
}

type imageInfo struct {
	registryRef string
	defaultEnv  map[string]string
}

type ReconcileComponent struct {
	controller2.ReconcilerHelper
	runtimeImages      map[string]imageInfo
	supervisor         *v1alpha2.Component
	onOpenShift        *bool
	dependentResources map[string]dependentResource
}

func (r *ReconcileComponent) PrimaryResourceType() runtime.Object {
	return new(v1alpha2.Component)
}

func (r *ReconcileComponent) SecondaryResourceTypes() []runtime.Object {
	return []runtime.Object{
		&corev1.Pod{},
		&appsv1.Deployment{},
		&corev1.Service{},
		&routev1.Route{},
	}
}

func (r *ReconcileComponent) IsPrimaryResourceValid(object runtime.Object) bool {
	// todo: implement
	return true
}

func (ReconcileComponent) asComponent(object runtime.Object) *v1alpha2.Component {
	return object.(*v1alpha2.Component)
}

func (r *ReconcileComponent) ResourceMetadata(object runtime.Object) controller2.ResourceMetadata {
	component := r.asComponent(object)
	return controller2.ResourceMetadata{
		Name:         component.Name,
		Status:       component.Status.Phase.String(),
		Created:      component.ObjectMeta.CreationTimestamp,
		ShouldDelete: !component.ObjectMeta.DeletionTimestamp.IsZero(),
	}
}

func (r *ReconcileComponent) Delete(object runtime.Object) (bool, error) {
	// todo: implement
	return false, nil
}

func (r *ReconcileComponent) CreateOrUpdate(object runtime.Object) (bool, error) {
	component := r.asComponent(object)
	installFn := r.installDevMode
	if v1alpha2.BuildDeploymentMode == component.Spec.DeploymentMode {
		installFn = r.installBuildMode
	}
	return installFn(component, component.Namespace)
}

func (r *ReconcileComponent) SetErrorStatus(object runtime.Object, e error) {
	r.setErrorStatus(r.asComponent(object), e)
}

func (r *ReconcileComponent) SetSuccessStatus(object runtime.Object) {
	component := r.asComponent(object)
	if component.Status.Phase != v1alpha2.ComponentReady {
		err := r.updateStatus(component, v1alpha2.ComponentReady)
		if err != nil {
			panic(err)
		}
	}
}

func (r *ReconcileComponent) Helper() controller2.ReconcilerHelper {
	return r.ReconcilerHelper
}

//Create the factory object
func (r *ReconcileComponent) createIfNeeded(instance *v1alpha2.Component, resourceType runtime.Object) (bool, error) {
	key, kind := getKeyAndKindFor(resourceType)
	resource, ok := r.dependentResources[key]
	if !ok {
		return false, fmt.Errorf("unknown dependent type %s", kind)
	}

	res, err := resource.fetch(resource, instance)
	if err != nil {
		// create the object
		obj, errBuildObject := resource.build(resource, instance)
		if errBuildObject != nil {
			return false, errBuildObject
		}
		if errors.IsNotFound(err) {
			err = r.Client.Create(context.TODO(), obj)
			if err != nil {
				r.ReqLogger.Error(err, "Failed to create new ", "kind", kind)
				return false, err
			}
			r.ReqLogger.Info("Created successfully", "kind", kind)
			return true, nil
		}
		r.ReqLogger.Error(err, "Failed to get", "kind", kind)
		return false, err
	} else {
		// if the resource defined an updater, use it to try to update the resource
		if resource.update != nil {
			return resource.update(res, resource, instance)
		}
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
	return controller2.NewGenericReconciler(r).Reconcile(request)
}

type installFnType func(component *v1alpha2.Component, namespace string) (bool, error)

func (r *ReconcileComponent) installAndUpdateStatus(component *v1alpha2.Component, request reconcile.Request, install installFnType) (reconcile.Result, error) {
	changed, err := install(component, request.Namespace)
	if err != nil {
		r.ReqLogger.Error(err, fmt.Sprintf("failed to install %s mode", component.Spec.DeploymentMode))
		r.setErrorStatus(component, err)
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: changed}, r.updateStatus(component, v1alpha2.ComponentReady)
}

func (r *ReconcileComponent) setInitialStatus(component *v1alpha2.Component, phase v1alpha2.ComponentPhase) error {
	if component.Generation == 1 && component.Status.Phase == "" {
		if err := r.updateStatus(component, phase); err != nil {
			r.ReqLogger.Info("Status update failed !")
			return err
		}
	}
	return nil
}
