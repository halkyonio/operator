package kubernetes

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func SetNamespaceAndOwnerReference(resource interface{}, component *v1alpha1.Component) {
	if obj, ok := resource.(metav1.Object); ok {
		obj.SetNamespace(component.ObjectMeta.GetNamespace())

		obj.SetOwnerReferences([]metav1.OwnerReference{
			*metav1.NewControllerRef(component, schema.GroupVersionKind{
				Group:   v1alpha1.SchemeGroupVersion.Group,
				Version: v1alpha1.SchemeGroupVersion.Version,
				Kind:    component.Kind,
			}),
		})
	}
}
