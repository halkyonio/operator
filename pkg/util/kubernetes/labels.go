package kubernetes

import "github.com/snowdrop/component-operator/pkg/apis/component/v1alpha1"

func PopulateK8sLabels(component *v1alpha1.Component, componentType string) map[string]string {
	labels := map[string]string{}
	labels[v1alpha1.RuntimeLabelKey] = component.Spec.Runtime
	labels[v1alpha1.RuntimeVersionLabelKey] = component.Spec.Version
	labels[v1alpha1.ComponentLabelKey] = componentType
	labels[v1alpha1.NameLabelKey] = component.Name
	labels[v1alpha1.ManagedByLabelKey] = "component-operator"
	return labels
}
