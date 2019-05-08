package kubernetes

import (
	"fmt"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"os"
)

const (
	// WatchNamespaceEnvVar is the constant for env variable WATCH_NAMESPACE
	// which is the namespace that the pod is currently running in.
	WatchNamespaceEnvVar = "WATCH_NAMESPACE"
)

// GetWatchNamespace returns the namespace the operator should be watching for changes
func GetWatchNamespace() (string, error) {
	ns, found := os.LookupEnv(WatchNamespaceEnvVar)
	if !found {
		return "", fmt.Errorf("%s must be set", WatchNamespaceEnvVar)
	}
	return ns, nil
}

func SetNamespace(resource interface{}, component *v1alpha2.Component) {
	if obj, ok := resource.(metav1.Object); ok {
		obj.SetNamespace(component.ObjectMeta.GetNamespace())
	}
}

func SetOwnerReferences(obj metav1.Object, component *v1alpha2.Component) {
	obj.SetOwnerReferences([]metav1.OwnerReference{
		*metav1.NewControllerRef(component, schema.GroupVersionKind{
			Group:   v1alpha2.GroupVersion.Group,
			Version: v1alpha2.GroupVersion.Version,
			Kind:    component.Kind,
		}),
	})
}

func SetNamespaceAndOwnerReference(resource interface{}, component *v1alpha2.Component) {
	if obj, ok := resource.(metav1.Object); ok {
		obj.SetNamespace(component.ObjectMeta.GetNamespace())
		SetOwnerReferences(obj, component)
	}
}
