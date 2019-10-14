package framework

import (
	"k8s.io/apimachinery/pkg/runtime"
)

type PrimaryResourceManager interface {
	PrimaryResourceType() runtime.Object
	WatchedSecondaryResourceTypes() []runtime.Object
	Delete(object Resource) error
	CreateOrUpdate(object Resource) error
	Helper() *K8SHelper
	GetDependentResourceFor(owner Resource, resourceType runtime.Object) (DependentResource, error)
	AddDependentResource(resource DependentResource)
	SetPrimaryResourceStatus(primary Resource, statuses []DependentResourceStatus) (needsUpdate bool)
	NewFrom(name string, namespace string, helper *K8SHelper) (Resource, error)
}
