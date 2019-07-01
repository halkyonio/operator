package component

import (
	"context"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

type namer func(*v1alpha2.Component) string
type labelsNamer func(*v1alpha2.Component) string
type builder func(dependentResource, *v1alpha2.Component) (runtime.Object, error)
type fetcher func(dependentResource, *v1alpha2.Component) (runtime.Object, error)
type updater func(runtime.Object, dependentResource, *v1alpha2.Component) (bool, error)

type dependentResource struct {
	name       namer
	labelsName labelsNamer
	build      builder
	fetch      fetcher
	update     updater
	prototype  runtime.Object
	kind       string
}

var defaultNamer namer = func(component *v1alpha2.Component) string {
	return component.Name
}
var buildNamer namer = func(component *v1alpha2.Component) string {
	return defaultNamer(component) + "-build"
}
var buildOrDevNamer = func(c *v1alpha2.Component) string {
	if v1alpha2.BuildDeploymentMode == c.Spec.DeploymentMode {
		return buildNamer(c)
	}
	return defaultNamer(c)
}

func (r *ReconcileComponent) genericFetcher(res dependentResource, c *v1alpha2.Component) (runtime.Object, error) {
	into := res.prototype.DeepCopyObject()
	if err := r.Client.Get(context.TODO(), types.NamespacedName{Name: res.name(c), Namespace: c.Namespace}, into); err != nil {
		r.ReqLogger.Info(res.kind + " doesn't exist")
		return nil, err
	}
	return into, nil
}

func (r *ReconcileComponent) addDependentResource(res runtime.Object, buildFn builder, nameFn namer) {
	r.addDependentResourceFull(res, buildFn, nameFn, nil, nil)
}

func (r *ReconcileComponent) addDependentResourceFull(res runtime.Object, buildFn builder, nameFn namer, labelsNameFn labelsNamer, updateFn updater) {
	key, kind := getKeyAndKindFor(res)
	r.dependentResources[key] = dependentResource{
		build:      buildFn,
		labelsName: labelsNameFn,
		update:     updateFn,
		name:       nameFn,
		prototype:  res,
		fetch:      r.genericFetcher,
		kind:       kind,
	}
}
