package framework

import (
	"k8s.io/apimachinery/pkg/runtime"
)

type PrimaryResourceManager interface {
	PrimaryResourceType() runtime.Object
	Delete(object Resource) error
	CreateOrUpdate(object Resource) error
	NewFrom(name string, namespace string) (Resource, error)
	GetDependentResourcesTypes() []DependentResource
	Helper() *K8SHelper
	SetHelper(helper *K8SHelper)
}
