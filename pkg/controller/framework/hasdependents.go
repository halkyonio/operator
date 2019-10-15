package framework

import (
	"context"
	"fmt"
	"halkyon.io/operator/pkg/util"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"
)

type HasDependents struct {
	dependents map[string]DependentResource
}

func NewHasDependents() *HasDependents {
	return &HasDependents{dependents: make(map[string]DependentResource, 7)}
}

func keyFor(resourceType runtime.Object) (key string) {
	t := reflect.TypeOf(resourceType)
	pkg := t.PkgPath()
	kind := util.GetObjectName(resourceType)
	key = pkg + "/" + kind
	return
}

func (b *HasDependents) CreateOrUpdate(helper *K8SHelper) error {
	for _, resource := range b.dependents {
		if e := createIfNeeded(resource, helper); e != nil {
			return e
		}
	}
	return nil
}

func createIfNeeded(res DependentResource, helper *K8SHelper) error {
	// if the resource specifies that it shouldn't be created, exit fast
	if !res.CanBeCreatedOrUpdated() {
		return nil
	}

	kind := util.GetObjectName(res.Prototype())
	object, err := res.Fetch(helper)
	if err != nil {
		if errors.IsNotFound(err) {
			// create the object
			obj, errBuildObject := res.Build()
			if errBuildObject != nil {
				return errBuildObject
			}

			// set controller reference if the resource should be owned
			if res.ShouldBeOwned() {
				// in most instances, resourceDefinedOwner == owner but some resources might want to return a different one
				resourceDefinedOwner := res.Owner()
				if e := controllerutil.SetControllerReference(resourceDefinedOwner.GetAPIObject().(v1.Object), obj.(v1.Object), helper.Scheme); e != nil {
					helper.ReqLogger.Error(err, "Failed to set owner", "owner", resourceDefinedOwner, "resource", res.Name())
					return e
				}
			}

			alreadyExists := false
			if err = helper.Client.Create(context.TODO(), obj); err != nil {
				// ignore error if it's to state that obj already exists
				alreadyExists = errors.IsAlreadyExists(err)
				if !alreadyExists {
					helper.ReqLogger.Error(err, "Failed to create new ", "kind", kind)
					return err
				}
			}
			if !alreadyExists {
				helper.ReqLogger.Info("Created successfully", "kind", kind, "name", obj.(v1.Object).GetName())
			}
			return nil
		}
		helper.ReqLogger.Error(err, "Failed to get", "kind", kind)
		return err
	} else {
		// if the resource defined an updater, use it to try to update the resource
		updated, err := res.Update(object)
		if err != nil {
			return err
		}
		if updated {
			if err = helper.Client.Update(context.TODO(), object); err != nil {
				helper.ReqLogger.Error(err, "Failed to update", "kind", kind)
			}
		}
		return err
	}
}

func (b *HasDependents) FetchUpdatedDependent(dependentType runtime.Object, helper *K8SHelper) (runtime.Object, error) {
	key := keyFor(dependentType)
	resource, ok := b.dependents[key]
	if !ok {
		return nil, fmt.Errorf("couldn't find any dependent resource of kind '%s'", util.GetObjectName(dependentType))
	}
	fetch, err := resource.Fetch(helper)
	if err != nil {
		return nil, err
	}
	return fetch, nil
}

func (b *HasDependents) AddDependentResource(resource DependentResource) {
	if resource.Owner() == nil {
		panic(fmt.Errorf("dependent resource %s must have an owner", resource.Name()))
	}
	prototype := resource.Prototype()
	key := keyFor(prototype)
	b.dependents[key] = resource
}

func (b *HasDependents) MustGetDependentResourceFor(owner Resource, resourceType runtime.Object) (resource DependentResource) {
	var e error
	if resource, e = b.GetDependentResourceFor(owner, resourceType); e != nil {
		panic(e)
	}
	return resource
}

func (b *HasDependents) GetDependentResourceFor(owner Resource, resourceType runtime.Object) (DependentResource, error) {
	resource, ok := b.dependents[keyFor(resourceType)]
	if !ok {
		return nil, fmt.Errorf("couldn't find any dependent resource of kind '%s'", util.GetObjectName(resourceType))
	}
	return resource.NewInstanceWith(owner), nil
}

func (b *HasDependents) WatchedSecondaryResourceTypes() []runtime.Object {
	watched := make([]runtime.Object, 0, len(b.dependents))
	for _, dep := range b.dependents {
		if dep.ShouldWatch() {
			watched = append(watched, dep.Prototype())
		}
	}
	return watched
}

func (b *HasDependents) ComputeStatus(current Resource, err error, helper *K8SHelper) (statuses []DependentResourceStatus, needsUpdate bool) {
	if err != nil {
		return statuses, current.SetErrorStatus(err)
	}
	statuses = b.areDependentResourcesReady(helper)
	msgs := make([]string, 0, len(statuses))
	for _, status := range statuses {
		if !status.Ready {
			msgs = append(msgs, fmt.Sprintf("%s => %s", status.DependentName, status.Message))
		}
	}
	if len(msgs) > 0 {
		msg := fmt.Sprintf("Waiting for the following resources: %s", strings.Join(msgs, " / "))
		helper.ReqLogger.Info(msg)
		// set the status but ignore the result since dependents are not ready, we do need to update and requeue in any case
		_ = current.SetInitialStatus(msg)
		current.SetNeedsRequeue(true)
		return statuses, true
	}

	return statuses, false
}

func (b *HasDependents) areDependentResourcesReady(helper *K8SHelper) (statuses []DependentResourceStatus) {
	statuses = make([]DependentResourceStatus, 0, len(b.dependents))
	for _, dependent := range b.dependents {
		if dependent.ShouldBeCheckedForReadiness() {
			fetched, err := b.FetchUpdatedDependent(dependent.Prototype(), helper)
			name := util.GetObjectName(dependent.Prototype())
			if err != nil {
				statuses = append(statuses, NewFailedDependentResourceStatus(name, err))
			} else {
				ready, message := dependent.IsReady(fetched)
				if !ready {
					statuses = append(statuses, NewFailedDependentResourceStatus(name, message))
				} else {
					statuses = append(statuses, NewReadyDependentResourceStatus(dependent.NameFrom(fetched), dependent.OwnerStatusField()))
				}
			}
		}
	}
	return statuses
}
