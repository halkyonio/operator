package kubernetes

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func SetNamespace(resource interface{}, component *v1alpha1.Component) {
	if obj, ok := resource.(metav1.Object); ok {
		obj.SetNamespace(component.ObjectMeta.GetNamespace())
	}
}

func SetOwnerReferences(obj metav1.Object, component *v1alpha1.Component) {
	obj.SetOwnerReferences([]metav1.OwnerReference{
		*metav1.NewControllerRef(component, schema.GroupVersionKind{
			Group:   v1alpha1.SchemeGroupVersion.Group,
			Version: v1alpha1.SchemeGroupVersion.Version,
			Kind:    component.Kind,
		}),
	})
}

func SetNamespaceAndOwnerReference(resource interface{}, component *v1alpha1.Component) {
	if obj, ok := resource.(metav1.Object); ok {
		obj.SetNamespace(component.ObjectMeta.GetNamespace())
		SetOwnerReferences(obj, component)
	}
}
