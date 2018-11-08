package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/runtime/scheme"
)


const (
	version   = "v1alpha1"
	groupName = "component.k8s.io"
)

var (
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
	// SchemeGroupVersion is the group version used to register these objects.
	SchemeGroupVersion = schema.GroupVersion{Group: groupName, Version: version}
)