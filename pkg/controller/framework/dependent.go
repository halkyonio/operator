package framework

import (
	"context"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

type DependentResource interface {
	Name() string
	Fetch(helper *K8SHelper) (runtime.Object, error)
	Build() (runtime.Object, error)
	Update(toUpdate runtime.Object) (bool, error)
	NewInstanceWith(owner Resource) DependentResource
	Owner() Resource
	Prototype() runtime.Object
	ShouldWatch() bool
	CanBeCreatedOrUpdated() bool
	ShouldBeOwned() bool
	IsReady(underlying runtime.Object) (ready bool, message string)
	OwnerStatusField() string
	ShouldBeCheckedForReadiness() bool
	NameFrom(underlying runtime.Object) string
}

type DependentResourceStatus struct {
	DependentName    string
	Ready            bool
	Message          string
	OwnerStatusField string
}

func NewFailedDependentResourceStatus(dependentName string, errorOrMessage interface{}) DependentResourceStatus {
	msg := ""
	switch errorOrMessage.(type) {
	case string:
		msg = errorOrMessage.(string)
	case error:
		msg = errorOrMessage.(error).Error()
	}
	return DependentResourceStatus{DependentName: dependentName, Ready: false, Message: msg}
}

func NewReadyDependentResourceStatus(dependentName string, fieldName string) DependentResourceStatus {
	return DependentResourceStatus{DependentName: dependentName, OwnerStatusField: fieldName, Ready: true}
}

type DependentResourceHelper struct {
	_owner     Resource
	_prototype runtime.Object
	_delegate  DependentResource
}

func (res DependentResourceHelper) IsReady(underlying runtime.Object) (ready bool, message string) {
	return true, ""
}

func (res DependentResourceHelper) ShouldBeCheckedForReadiness() bool {
	return false
}

func (res DependentResourceHelper) OwnerStatusField() string {
	return ""
}

func (res DependentResourceHelper) NameFrom(underlying runtime.Object) string {
	return res.Name()
}

func (res DependentResourceHelper) ShouldWatch() bool {
	return true
}

func (res DependentResourceHelper) ShouldBeOwned() bool {
	return true
}

func (res DependentResourceHelper) CanBeCreatedOrUpdated() bool {
	return true
}

func NewDependentResource(primaryResourceType runtime.Object, owner Resource) *DependentResourceHelper {
	return &DependentResourceHelper{_prototype: primaryResourceType, _owner: owner}
}

func (res *DependentResourceHelper) SetDelegate(delegate DependentResource) {
	res._delegate = delegate
}

func (res DependentResourceHelper) Name() string {
	return DefaultDependentResourceNameFor(res.Owner())
}

func (res DependentResourceHelper) Fetch(helper *K8SHelper) (runtime.Object, error) {
	delegate := res._delegate
	into := delegate.Prototype()
	if err := helper.Client.Get(context.TODO(), types.NamespacedName{Name: delegate.Name(), Namespace: delegate.Owner().GetNamespace()}, into); err != nil {
		return nil, err
	}
	return into, nil
}

func (res DependentResourceHelper) Owner() Resource {
	return res._owner
}

func (res DependentResourceHelper) Prototype() runtime.Object {
	return res._prototype.DeepCopyObject()
}

func DefaultDependentResourceNameFor(owner Resource) string {
	return owner.GetName()
}
