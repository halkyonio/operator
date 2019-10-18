package component

import (
	routev1 "github.com/openshift/api/route/v1"
	"halkyon.io/operator/pkg/controller/framework"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type route struct {
	base
}

func newRoute(owner framework.Resource) route {
	dependent := newBaseDependent(&routev1.Route{}, owner)
	r := route{base: dependent}
	dependent.SetDelegate(r)
	return r
}

//buildRoute returns the route resource
func (res route) Build() (runtime.Object, error) {
	c := res.ownerAsComponent()
	ls := getAppLabels(DeploymentName(c))
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
	return framework.IsTargetClusterRunningOpenShift()
}

func (res route) CanBeCreatedOrUpdated() bool {
	return res.ownerAsComponent().Spec.ExposeService && res.ShouldWatch()
}
