package component

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
)

func (r *ReconcileComponent) getEnvAsMap(component v1alpha2.ComponentSpec) (map[string]string, error) {
	envs := component.Envs
	tmpEnvVar := make(map[string]string)

	for _, v := range envs {
		tmpEnvVar[v.Name] = v.Value
	}

	image, err := r.getImageInfo(component)
	if err != nil {
		return map[string]string{}, err
	}

	// Check if Component EnvVar contains the defaults
	for k, v := range image.defaultEnv {
		if _, ok := tmpEnvVar[k]; !ok {
			tmpEnvVar[k] = v
		}
	}

	return tmpEnvVar, nil
}

func (r *ReconcileComponent) populateEnvVar(component *v1alpha2.Component) {
	tmpEnvVar, err := r.getEnvAsMap(component.Spec)
	if err != nil {
		panic(err)
	}

	// Convert Map to Slice
	newEnvVars := make([]v1alpha2.Env, 0, len(tmpEnvVar))
	for k, v := range tmpEnvVar {
		newEnvVars = append(newEnvVars, v1alpha2.Env{Name: k, Value: v})
	}

	// Store result
	component.Spec.Envs = newEnvVars
}

//getAppLabels returns an string map with the labels which wil be associated to the kubernetes/ocp resource which will be created and managed by this operator
func (r *ReconcileComponent) getAppLabels(name string) map[string]string {
	return map[string]string{
		"app":          name,
		"component_cr": name,
		"deployment":   name,
	}
}

//Check if the mandatory specs are filled
func (r *ReconcileComponent) hasMandatorySpecs(instance *v1alpha2.Component) bool {
	// TODO
	return true
}

func (r *ReconcileComponent) PopulateK8sLabels(component *v1alpha2.Component, componentType string) map[string]string {
	labels := map[string]string{}
	labels[v1alpha2.RuntimeLabelKey] = component.Spec.Runtime
	labels[v1alpha2.RuntimeVersionLabelKey] = component.Spec.Version
	labels[v1alpha2.ComponentLabelKey] = componentType
	labels[v1alpha2.NameLabelKey] = component.Name
	labels[v1alpha2.ManagedByLabelKey] = "component-operator"
	return labels
}
