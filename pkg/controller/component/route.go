package component

import (
	routev1 "github.com/openshift/api/route/v1"
	"halkyon.io/operator/pkg/controller"
	"halkyon.io/operator/pkg/controller/framework"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type route struct {
	base
	reconciler *ReconcileComponent // todo: remove
}

func (res route) NewInstanceWith(owner framework.Resource) framework.DependentResource {
	return newOwnedRoute(res.reconciler, owner)
}

func newRoute(reconciler *ReconcileComponent) route {
	return newOwnedRoute(reconciler, nil)
}

func newOwnedRoute(reconciler *ReconcileComponent, owner framework.Resource) route {
	dependent := newBaseDependent(&routev1.Route{}, owner)
	r := route{base: dependent, reconciler: reconciler}
	dependent.SetDelegate(r)
	return r
}

//buildRoute returns the route resource
func (res route) Build() (runtime.Object, error) {
	c := res.ownerAsComponent()
	ls := getAppLabels(controller.DeploymentName(c))
	route := &routev1.Route{
		ObjectMeta: v1.ObjectMeta{
			Name:      res.Name(),
			Namespace: c.Namespace,
			Labels:    ls,
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: c.Name,
			},
		},
	}

	return route, nil
}

func (res route) ShouldWatch() bool {
	return res.reconciler.IsTargetClusterRunningOpenShift()
}

func (res route) CanBeCreatedOrUpdated() bool {
	return res.ownerAsComponent().Spec.ExposeService && res.ShouldWatch()
}
