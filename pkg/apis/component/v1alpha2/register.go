package v1alpha2

import (
	servicecatalog "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	deploymentconfig "github.com/openshift/api/apps/v1"
	build "github.com/openshift/api/build/v1"
	image "github.com/openshift/api/image/v1"
	route "github.com/openshift/api/route/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	version   = "v1alpha2"
	groupName = "component.k8s.io"
)

var (
	GroupName = groupName
	// SchemeGroupVersion is the group version used to register these objects.
	GroupVersion  = schema.GroupVersion{Group: GroupName, Version: version}
	schemeBuilder = runtime.NewSchemeBuilder(addKnownTypes,
		deploymentconfig.Install,
		image.Install,
		route.Install,
		servicecatalog.AddToScheme,
		build.Install)
	// Install is a function which adds this version to a scheme
	Install = schemeBuilder.AddToScheme
)

// Adds the list of known types to api.Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion,
		&Component{},
		&ComponentList{},
		&Link{},
		&LinkList{},
		&Service{},
		&ServiceList{},
	)
	metav1.AddToGroupVersion(scheme, GroupVersion)
	return nil
}
