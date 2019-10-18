package framework

import (
	"fmt"
	"halkyon.io/operator/pkg/util"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Resource interface {
	v1.Object
	runtime.Object
	NeedsRequeue() bool
	SetNeedsRequeue(requeue bool)
	GetStatusAsString() string
	ShouldDelete() bool
	SetErrorStatus(err error) bool
	SetSuccessStatus(statuses []DependentResourceStatus, msg string) bool
	SetInitialStatus(msg string) bool
	ComputeStatus(err error, helper *K8SHelper) (needsUpdate bool)
	CheckValidity() error
	Init() bool
	GetAPIObject() runtime.Object
	AddDependentResource(resource DependentResource)
	FetchUpdatedDependent(dependentType runtime.Object, helper *K8SHelper) (runtime.Object, error)
	FetchAndInit(name, namespace string) (Resource, error)
	PrimaryResourceType() runtime.Object
	Delete() error
	CreateOrUpdate() error
	GetDependentResourcesTypes() []DependentResource
}

func HasChangedFromStatusUpdate(status interface{}, statuses []DependentResourceStatus, msg string) (changed bool, updatedMsg string) {
	updatedMsg = msg
	for _, s := range statuses {
		changed = changed || util.MustSetNamedStringField(status, s.OwnerStatusField, s.DependentName)
		if changed {
			updatedMsg = fmt.Sprintf("%s: '%s' changed to '%s'", msg, s.OwnerStatusField, s.DependentName)
		}
	}
	return changed, updatedMsg
}
