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
	GetStatusAsString() string
	ShouldDelete() bool
	SetErrorStatus(err error) bool
	SetInitialStatus(msg string) bool
	ComputeStatus() (needsUpdate bool)
	CheckValidity() error
	Init() bool
	GetAPIObject() runtime.Object
	FetchAndCreateNew(name, namespace string) (Resource, error)
	PrimaryResourceType() runtime.Object
	Delete() error
	CreateOrUpdate() error
	GetDependentResourcesTypes() map[string]DependentResource
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
