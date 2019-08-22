package component

import (
	"halkyon.io/operator/pkg/apis/halkyon/v1beta1"
	"os"
)

const (
	RegistryAddressEnvVar = "REGISTRY_ADDRESS"
	BaseS2iImage = "BASE_S2I_IMAGE"
)

func (r *ReconcileComponent) getEnvAsMap(component v1beta1.ComponentSpec) (map[string]string, error) {
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

func (r *ReconcileComponent) populateEnvVar(component *v1beta1.Component) {
	tmpEnvVar, err := r.getEnvAsMap(component.Spec)
	if err != nil {
		panic(err)
	}

	// Convert Map to Slice
	newEnvVars := make([]v1beta1.Env, 0, len(tmpEnvVar))
	for k, v := range tmpEnvVar {
		newEnvVars = append(newEnvVars, v1beta1.Env{Name: k, Value: v})
	}

	// Store result
	component.Spec.Envs = newEnvVars
}

//getAppLabels returns a string map with the Application labels which will be associated to the kubernetes/ocp resource created and managed by this operator
func getAppLabels(name string) map[string]string {
	return map[string]string{
		"app":          name,
		"component_cr": name,
		"deployment":   name,
	}
}

//getBuildLabels returns a string map with the Build labels which wil be associated to the kubernetes/ocp resource created and managed by this operator
func getBuildLabels(name string) map[string]string {
	return map[string]string{
		"build":        name,
		"component_cr": name,
	}
}

//Check if the mandatory specs are filled
func (r *ReconcileComponent) hasMandatorySpecs(instance *v1beta1.Component) bool {
	// TODO
	return true
}

func (r *ReconcileComponent) PopulateK8sLabels(component *v1beta1.Component, componentType string) map[string]string {
	labels := map[string]string{}
	labels[v1beta1.RuntimeLabelKey] = component.Spec.Runtime
	labels[v1beta1.RuntimeVersionLabelKey] = component.Spec.Version
	labels[v1beta1.ComponentLabelKey] = componentType
	labels[v1beta1.NameLabelKey] = component.Name
	labels[v1beta1.ManagedByLabelKey] = "halkyon-operator"
	return labels
}

func (r *ReconcileComponent) baseImage(c *v1beta1.Component) string {
	if c.Spec.BuildConfig.BaseImage != "" {
		return c.Spec.BuildConfig.BaseImage
	} else {
		baseImage, found := os.LookupEnv("BaseS2iImage")
		if found {
			return baseImage
		} else {
			// We return the default image
			return "registry.access.redhat.com/redhat-openjdk-18/openjdk18-openshift"
		}
	}
}

func (r *ReconcileComponent) gitRevision(c *v1beta1.Component) string {
	if c.Spec.BuildConfig.Ref == "" {
		return "master"
	} else {
		return c.Spec.BuildConfig.Ref
	}
}

func (r *ReconcileComponent) dockerImageURL(c *v1beta1.Component) string {
	// Try to find the registry env var
	registry, found := os.LookupEnv(RegistryAddressEnvVar)
	if found {
		return registry + "/" + c.Namespace + "/" + c.Name
	}
	// Revert to default values if no ENV var is set
	if r.IsTargetClusterRunningOpenShift() {
		if r.OpenShiftVersion() == 4 {
			return "image-registry.openshift-image-registry.svc:5000/" + c.Namespace + "/" + c.Name
		} else {
			return "docker-registry.default.svc:5000/" + c.Namespace + "/" + c.Name
		}
	} else {
		return "kube-registry.kube-system.svc:5000/" + c.Namespace + "/" + c.Name
	}
}
