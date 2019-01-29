package apis

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha1"

	servicecatalog "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	deploymentconfig "github.com/openshift/api/apps/v1"
	image "github.com/openshift/api/image/v1"
	route "github.com/openshift/api/route/v1"
	build "github.com/openshift/api/build/v1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes,
		v1alpha1.SchemeBuilder.AddToScheme,
		deploymentconfig.AddToScheme,
		image.AddToScheme,
		route.AddToScheme,
		servicecatalog.AddToScheme,
		build.AddToScheme)
}
