package component

import (
	"context"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
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

func (r *ReconcileComponent) initDependentResources() {
	r.addDependentResource(&corev1.PersistentVolumeClaim{}, r.buildPVC, func(c *v1alpha2.Component) string {
		specified := c.Spec.Storage.Name
		if len(specified) > 0 {
			return specified
		}
		return "m2-data-" + c.Name
	})
	r.addDependentResource(&appsv1.Deployment{},
		func(res dependentResource, c *v1alpha2.Component) (object runtime.Object, e error) {
			if v1alpha2.BuildDeploymentMode == c.Spec.DeploymentMode {
				if err := r.setInitialStatus(c, v1alpha2.ComponentBuilding); err != nil {
					return nil, err
				}
				return r.createBuildDeployment(res, c)
			}
			if err := r.setInitialStatus(c, v1alpha2.ComponentPending); err != nil {
				return nil, err
			}
			return r.buildDevDeployment(res, c)
		}, buildOrDevNamer)
	r.addDependentResourceFull(&corev1.Service{}, r.buildService, defaultNamer, buildOrDevNamer, r.updateServiceSelector)
	r.addDependentResource(&corev1.ServiceAccount{}, r.buildServiceAccount, func(c *v1alpha2.Component) string {
		return serviceAccountName
	})
	r.addDependentResource(&routev1.Route{}, r.buildRoute, defaultNamer)
	r.addDependentResource(&v1beta1.Ingress{}, r.buildIngress, defaultNamer)
	taskNamer := func(c *v1alpha2.Component) string {
		return taskS2iBuildahPushName
	}
	r.addDependentResource(&v1alpha1.Task{}, r.buildTaskS2iBuildahPush, taskNamer)
	r.addDependentResource(&v1alpha1.TaskRun{}, r.buildTaskRunS2iBuildahPush, defaultNamer)
}
