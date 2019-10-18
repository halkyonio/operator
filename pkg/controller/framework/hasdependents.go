package framework

import (
	"fmt"
	"halkyon.io/operator/pkg/util"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
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

func (b *HasDependents) CreateOrUpdateDependents(helper *K8SHelper) error {
	for _, dep := range b.dependents {
		if e := dep.CreateOrUpdate(helper); e != nil {
			return e
		}
	}
	return nil
}

func (b *HasDependents) FetchAndInitNewResource(name string, namespace string, toInit Resource) (Resource, error) {
	toInit.SetName(name)
	toInit.SetNamespace(namespace)
	resourceType := toInit.GetAPIObject()
	helper := GetHelperFor(resourceType)
	_, err := helper.Fetch(name, namespace, resourceType)
	if err != nil {
		return toInit, err
	}
	resourcesTypes := toInit.GetDependentResourcesTypes()
	for _, rType := range resourcesTypes {
		toInit.AddDependentResource(rType.NewInstanceWith(toInit))
	}
	return toInit, err
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
