package kubernetes

import "github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"

func PopulateK8sLabels(component *v1alpha2.Component, componentType string) map[string]string {
	labels := map[string]string{}
	labels[v1alpha2.RuntimeLabelKey] = component.Spec.Runtime
	labels[v1alpha2.RuntimeVersionLabelKey] = component.Spec.Version
	labels[v1alpha2.ComponentLabelKey] = componentType
	labels[v1alpha2.NameLabelKey] = component.Name
	labels[v1alpha2.ManagedByLabelKey] = "component-operator"
	return labels
}
