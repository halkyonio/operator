package component

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
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

	image["java"] = "quay.io/snowdrop/spring-boot-s2i"
	image["nodejs"] = "nodeshift/centos7-s2i-nodejs:10.x"
	image["supervisord"] = "quay.io/snowdrop/supervisord"
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

func (r *ReconcileComponent) createTypeImage(dockerImage bool, name string, tag string, repo string, annotationCmd bool) v1alpha2.Image {
	return v1alpha2.Image{
		DockerImage:    dockerImage,
		Name:           name,
		Repo:           repo,
		AnnotationCmds: annotationCmd,
		Tag:            tag,
	}
}

func (r *ReconcileComponent) getSupervisordImage() []v1alpha2.Image {
	return []v1alpha2.Image{
		r.createTypeImage(true, "copy-supervisord", "latest", image["supervisord"], true),
	}
}

//Check if the mandatory specs are filled
func (r *ReconcileComponent) hasMandatorySpecs(instance *v1alpha2.Component) bool {
	// TODO
	return true
}

