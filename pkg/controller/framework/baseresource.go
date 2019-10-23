package framework

import (
	"fmt"
	"halkyon.io/operator/pkg/util"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"strings"
)

type BaseResource struct {
	dependents          map[string]DependentResource
	requeue             bool
	helper              *K8SHelper
	primaryResourceType runtime.Object
}

func (b *BaseResource) PrimaryResourceType() runtime.Object {
	return b.primaryResourceType
}

func (b *BaseResource) Helper() *K8SHelper {
	if b.helper == nil {
		b.helper = GetHelperFor(b.primaryResourceType)
	}
	return b.helper
}

func (b *BaseResource) SetNeedsRequeue(requeue bool) {
	b.requeue = requeue
}

func (b *BaseResource) NeedsRequeue() bool {
	return b.requeue
}

func NewHasDependents(primary runtime.Object) *BaseResource {
	return &BaseResource{dependents: make(map[string]DependentResource, 15), primaryResourceType: primary}
}

func keyFor(resourceType runtime.Object) (key string) {
	t := reflect.TypeOf(resourceType)
	pkg := t.PkgPath()
	kind := util.GetObjectName(resourceType)
	key = pkg + "/" + kind
	return
}

func (b *BaseResource) CreateOrUpdateDependents() error {
	for _, dep := range b.dependents {
		if e := dep.CreateOrUpdate(b.Helper()); e != nil {
			return e
		}
	}
	return nil
}

func (b *BaseResource) FetchAndInitNewResource(name string, namespace string, toInit Resource) (Resource, error) {
	toInit.SetName(name)
	toInit.SetNamespace(namespace)
	resourceType := toInit.GetAPIObject()
	_, err := b.Helper().Fetch(name, namespace, resourceType)
	if err != nil {
		return toInit, err
	}
	return toInit, err
}

func (b *BaseResource) FetchUpdatedDependent(dependentType runtime.Object) (runtime.Object, error) {
	key := keyFor(dependentType)
	resource, ok := b.dependents[key]
	if !ok {
		return nil, fmt.Errorf("couldn't find any dependent resource of kind '%s'", util.GetObjectName(dependentType))
	}
	fetch, err := resource.Fetch(b.Helper())
	if err != nil {
		return nil, err
	}
	return fetch, nil
}

func (b *BaseResource) GetDependentResourcesTypes() map[string]DependentResource {
	return b.dependents
}

func (b *BaseResource) AddDependentResource(resources ...DependentResource) {
	for _, resource := range resources {
		if resource.Owner() == nil {
			panic(fmt.Errorf("dependent resource %s must have an owner", resource.Name()))
		}
		prototype := resource.Prototype()
		key := keyFor(prototype)
		b.dependents[key] = resource
	}
}

func (b *BaseResource) ComputeStatus(current Resource) (statuses []DependentResourceStatus, notReadyWantsUpdate bool) {
	statuses = b.areDependentResourcesReady()
	msgs := make([]string, 0, len(statuses))
	for _, status := range statuses {
		if !status.Ready {
			msgs = append(msgs, fmt.Sprintf("%s => %s", status.DependentName, status.Message))
		}
	}
	if len(msgs) > 0 {
		msg := fmt.Sprintf("Waiting for the following resources: %s", strings.Join(msgs, " / "))
		b.Helper().ReqLogger.Info(msg)
		// set the status but ignore the result since dependents are not ready, we do need to update and requeue in any case
		_ = current.SetInitialStatus(msg)
		b.SetNeedsRequeue(true)
		return statuses, true
	}

	return statuses, false
}

func (b *BaseResource) areDependentResourcesReady() (statuses []DependentResourceStatus) {
	statuses = make([]DependentResourceStatus, 0, len(b.dependents))
	for _, dependent := range b.dependents {
		if dependent.ShouldBeCheckedForReadiness() {
			objectType := dependent.Prototype()
			fetched, err := b.FetchUpdatedDependent(objectType)
			name := util.GetObjectName(objectType)
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
