package component

import (
	routev1 "github.com/openshift/api/route/v1"
	component "halkyon.io/api/component/v1beta1"
	"halkyon.io/api/v1beta1"
	"halkyon.io/operator-framework"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type route struct {
	base
}

var _ framework.DependentResource = &route{}

func newRoute(owner v1beta1.HalkyonResource) route {
	config := framework.NewConfig(routev1.GroupVersion.WithKind("Route"), owner.GetNamespace())
	config.Watched = framework.IsTargetClusterRunningOpenShift()
	config.CreatedOrUpdated = owner.(*component.Component).Spec.ExposeService && config.Watched
	return route{base: newConfiguredBaseDependent(owner, config)}
}

//buildRoute returns the route resource
func (res route) Build(empty bool) (runtime.Object, error) {
	route := &routev1.Route{}
	if !empty {
		c := res.ownerAsComponent()
		ls := getAppLabels(DeploymentName(c))
		route.ObjectMeta = v1.ObjectMeta{
			Name:      res.Name(),
			Namespace: c.Namespace,
			Labels:    ls,
		}
		route.Spec = routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: c.Name,
			},
		}
	}

	return route, nil
}
