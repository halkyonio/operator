package controller

import (
	"fmt"
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
	IsValid() bool
	Clone() Resource
	GetAPIObject() runtime.Object
	SetAPIObject(object runtime.Object)
}

func hasChangedFromStatusUpdate(status interface{}, statuses []DependentResourceStatus, msg string) (changed bool, updatedMsg string) {
	updatedMsg = msg
	for _, s := range statuses {
		changed = changed || MustSetNamedStringField(status, s.OwnerStatusField, s.DependentName)
		if changed {
			updatedMsg = fmt.Sprintf("%s: '%s' changed to '%s'", msg, s.OwnerStatusField, s.DependentName)
		}
	}
	return changed, updatedMsg
}
