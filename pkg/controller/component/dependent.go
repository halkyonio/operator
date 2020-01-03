package component

import (
	"halkyon.io/api/component/v1beta1"
	v1beta12 "halkyon.io/api/v1beta1"
	"halkyon.io/operator-framework"
	"halkyon.io/operator-framework/util"
	"k8s.io/apimachinery/pkg/runtime"
)

type base struct {
	*framework.BaseDependentResource
}

func (res base) Name() string {
	return framework.DefaultDependentResourceNameFor(res.Owner())
}

func (res base) Build(empty bool) (runtime.Object, error) {
	panic("implement me")
}

func (res base) NameFrom(underlying runtime.Object) string {
	return framework.DefaultNameFrom(res, underlying)
}

func (res base) Fetch(helper *framework.K8SHelper) (runtime.Object, error) {
	return framework.DefaultFetcher(res, helper)
}

func (res base) IsReady(underlying runtime.Object) (ready bool, message string) {
	return framework.DefaultIsReady(underlying)
}

func (res base) Update(toUpdate runtime.Object) (bool, error) {
	return false, nil
}

func newBaseDependent(primaryResourceType runtime.Object, owner v1beta12.HalkyonResource) base {
	gvk := util.GetGVKFor(primaryResourceType, owner.(*Component).Helper().Scheme)
	return base{framework.NewBaseDependentResource(owner, gvk)}
}

func newConfiguredBaseDependent(owner v1beta12.HalkyonResource, config framework.DependentResourceConfig) base {
	return base{framework.NewConfiguredBaseDependentResource(owner, config)}
}

func asHalkyonComponent(res v1beta12.HalkyonResource) *v1beta1.Component {
	return res.(*Component).Component
}

func ownerAsHalkyonComponent(res framework.DependentResource) *v1beta1.Component {
	return asHalkyonComponent(res.Owner())
}

func (res base) ownerAsComponent() *v1beta1.Component {
	return ownerAsHalkyonComponent(res)
}

func (res base) asComponent(object runtime.Object) *Component {
	return object.(*Component)
}
