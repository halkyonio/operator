package component

import (
	"halkyon.io/api/component/v1beta1"
	beta1 "halkyon.io/api/v1beta1"
	"halkyon.io/operator-framework"
	"halkyon.io/operator-framework/util"
	"k8s.io/apimachinery/pkg/runtime"
)

type base struct {
	*framework.BaseDependentResource
	NameFn func() string
}

func (res base) Name() string {
	if res.NameFn != nil {
		return res.NameFn()
	}
	return framework.DefaultDependentResourceNameFor(res.Owner())
}

func (res base) Build(empty bool) (runtime.Object, error) {
	panic("implement me")
}

func (res base) NameFrom(underlying runtime.Object) string {
	return framework.DefaultNameFrom(res, underlying)
}

func (res base) Fetch() (runtime.Object, error) {
	return framework.DefaultFetcher(res)
}

func (res base) GetCondition(_ runtime.Object, err error) *beta1.DependentCondition {
	return framework.DefaultGetConditionFor(res, err)
}

func (res base) Update(toUpdate runtime.Object) (bool, error) {
	return false, nil
}

func newBaseDependent(primaryResourceType runtime.Object, owner *v1beta1.Component) base {
	gvk := util.GetGVKFor(primaryResourceType, framework.Helper.Scheme)
	return base{BaseDependentResource: framework.NewBaseDependentResource(owner, gvk)}
}

func newConfiguredBaseDependent(owner *v1beta1.Component, config framework.DependentResourceConfig) base {
	return base{BaseDependentResource: framework.NewConfiguredBaseDependentResource(owner, config)}
}

func (res base) ownerAsComponent() *v1beta1.Component {
	return res.Owner().(*v1beta1.Component)
}

func (res base) asComponent(object runtime.Object) *Component {
	return object.(*Component)
}
