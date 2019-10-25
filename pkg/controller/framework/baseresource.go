package framework

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"strings"
)

type BaseResource struct {
	dependents          []DependentResource
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
	return &BaseResource{dependents: make([]DependentResource, 0, 15), primaryResourceType: primary}
}

func (b *BaseResource) CreateOrUpdateDependents() error {
	for _, dep := range b.dependents {
		if e := dep.CreateOrUpdate(b.Helper()); e != nil {
			return fmt.Errorf("failed to create or update '%s' %s: %s", dep.Name(), dep.GetTypeName(), e.Error())
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

func (b *BaseResource) FetchUpdatedDependent(dependentType string) (runtime.Object, error) {
	var dependent DependentResource
	for _, d := range b.dependents {
		if d.GetTypeName() == dependentType {
			dependent = d
			break
		}
	}
	if dependent == nil {
		return nil, fmt.Errorf("couldn't find any dependent resource of kind '%s'", dependentType)
	}
	fetch, err := dependent.Fetch(b.Helper())
	if err != nil {
		return nil, err
	}
	return fetch, nil
}

func (b *BaseResource) GetDependentResourcesTypes() []DependentResource {
	return b.dependents
}

// AddDependentResource adds dependent resources to this base resource, keeping the order in which they are added, it is
// therefore possible to create dependent resources in a specific order since they are created in the same order as specified here
func (b *BaseResource) AddDependentResource(resources ...DependentResource) {
	for _, dependent := range resources {
		if dependent.Owner() == nil {
			panic(fmt.Errorf("dependent resource %s must have an owner", dependent.Name()))
		}
		b.dependents = append(b.dependents, dependent)
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
			name := dependent.GetTypeName()
			fetched, err := b.FetchUpdatedDependent(name)
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
