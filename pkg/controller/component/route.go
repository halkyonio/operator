package component

import (
	routev1 "github.com/openshift/api/route/v1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

//buildRoute returns the route resource
func (r *ReconcileComponent) buildRoute(c *v1alpha2.Component) *routev1.Route {
	ls := r.getAppLabels(c.Name)
	route := &routev1.Route{
		TypeMeta: v1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Route",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      c.Name,
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

	// Set Component instance as the owner and controller
	controllerutil.SetControllerReference(c, route, r.scheme)
	return route
}
