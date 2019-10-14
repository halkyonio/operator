package framework

import (
	"fmt"
	"halkyon.io/operator/pkg/util"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
)

type HasDependents struct {
	dependents map[string]DependentResource
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
		if e := CreateIfNeeded(resource, helper); e != nil {
			return e
		}
	}
	return nil
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

/*func (b *HasDependents) computeStatus(current Resource, err error) (needsUpdate bool) {
	if err != nil {
		return current.SetErrorStatus(err)
	}
	statuses := b.areDependentResourcesReady(current)
	msgs := make([]string, 0, len(statuses))
	for _, status := range statuses {
		if !status.Ready {
			msgs = append(msgs, fmt.Sprintf("%s => %s", status.DependentName, status.Message))
		}
	}
	if len(msgs) > 0 {
		msg := fmt.Sprintf("Waiting for the following resources: %s", strings.Join(msgs, " / "))
		b.ReqLogger.Info(msg)
		// set the status but ignore the result since dependents are not ready, we do need to update and requeue in any case
		_ = current.SetInitialStatus(msg)
		current.SetNeedsRequeue(true)
		return true
	}

	return b.factory().SetPrimaryResourceStatus(current, statuses, nil)
}

func (b *HasDependents) areDependentResourcesReady(resource Resource) (statuses []DependentResourceStatus) {
	statuses = make([]DependentResourceStatus, 0, len(b.dependents))
	for _, dependent := range b.dependents {
		// make sure owner is set:
		dependent = dependent.NewInstanceWith(resource)
		b.dependents[keyFor(dependent.Prototype())] = dependent

		if dependent.ShouldBeCheckedForReadiness() {
			fetched, err := dependent.Fetch(b.Helper())
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
*/
