package component

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	corev1 "k8s.io/api/core/v1"
)

var (
	defaultEnvVar = make(map[string]string)
	image         = make(map[string]string)
)

func init() {
	defaultEnvVar["JAVA_APP_DIR"] = "/deployment"
	defaultEnvVar["JAVA_DEBUG"] = "\"false\""
	defaultEnvVar["JAVA_DEBUG_PORT"] = "\"5005\""
	defaultEnvVar["JAVA_APP_JAR"] = "app.jar"
	// defaultEnvVar["CMDS"] = "run-java:/usr/local/s2i/run;run-node:/usr/libexec/s2i;compile-java:/usr/local/s2i/assemble;build:/deployments/buildapp"
}

func (r *ReconcileComponent) populateEnvVar(component *v1alpha2.Component) {
	envs := component.Spec.Envs
	tmpEnvVar := make(map[string]string)

	// Convert Slice to Map
	for i := 0; i < len(envs); i += 1 {
		tmpEnvVar[envs[i].Name] = envs[i].Value
	}

	// Check if Component EnvVar contains the defaults
	for k, _ := range defaultEnvVar {
		if tmpEnvVar[k] == "" {
			tmpEnvVar[k] = defaultEnvVar[k]
		}
	}

	// Convert Map to Slice
	newEnvVars := []v1alpha2.Env{}
	for k, v := range tmpEnvVar {
		newEnvVars = append(newEnvVars, v1alpha2.Env{Name: k, Value: v})
	}

	// Store result
	component.Spec.Envs = newEnvVars
}

func (r *ReconcileComponent) populatePodEnvVar(component *v1alpha2.Component) *[]corev1.EnvVar{
	envs := component.Spec.Envs
	tmpEnvVar := make(map[string]string)

	// Convert Slice to Map
	for i := 0; i < len(envs); i += 1 {
		tmpEnvVar[envs[i].Name] = envs[i].Value
	}

	// Check if Component EnvVar contains the defaults
	for k, _ := range defaultEnvVar {
		if tmpEnvVar[k] == "" {
			tmpEnvVar[k] = defaultEnvVar[k]
		}
	}

	// Convert Map to Slice
	newEnvVars := []corev1.EnvVar{}
	for k, v := range tmpEnvVar {
		newEnvVars = append(newEnvVars, corev1.EnvVar{Name: k, Value: v})
	}

	return &newEnvVars
}

//getAppLabels returns an string map with the labels which wil be associated to the kubernetes/ocp resource which will be created and managed by this operator
func (r *ReconcileComponent) getAppLabels(name string) map[string]string {
	return map[string]string{
		"app": name,
		"component_cr": name,
		"deploymentconfig": name,
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

