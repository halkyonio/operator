package kubernetes

import "github.com/snowdrop/component-operator/pkg/apis/component/v1alpha1"

func EnrichComponentWithLabels(component *v1alpha1.Component) {
	labels := map[string]string{}
	labels[v1alpha1.RuntimeLabelKey] = component.Spec.Runtime
	labels[v1alpha1.RuntimeVersionLabelKey] = component.Spec.Version
	labels[v1alpha1.ComponentLabelKey] = "backend"
	labels[v1alpha1.NameLabelKey] = component.Name
	labels[v1alpha1.ManagedByLabelKey] = "component-operator"
	component.Labels = labels
}
