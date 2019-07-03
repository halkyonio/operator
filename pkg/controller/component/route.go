package component

import (
	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type route struct {
	base
	reconciler *ReconcileComponent // todo: remove
}

func newRoute(reconciler *ReconcileComponent) route {
	return route{base: newBaseDependent(&routev1.Route{}), reconciler: reconciler}
}

//buildRoute returns the route resource
func (res route) Build() (runtime.Object, error) {
	c := res.ownerAsComponent()
	ls := getAppLabels(c.Name)
	route := &routev1.Route{
		TypeMeta: v1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Route",
		},
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
	return res.reconciler.isTargetClusterRunningOpenShift()
}
