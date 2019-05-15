package v1alpha2

import (
	"fmt"
	"github.com/snowdrop/component-operator/pkg/util"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

// ComponentSpec defines the desired state of Component
// +k8s:openapi-gen=true
type ComponentSpec struct {
	// DeploymentMode indicates the strategy to be adopted to install the resources into a namespace
	// and next to create a pod. 2 strategies are currently supported; inner and outer loop
	// where outer loop refers to a build of the code and the packaging of the application into a container's image
	// while the inner loop will install a pod's running a supervisord daemon used to trigger actions such as : assemble, run, ...
	DeploymentMode string `json:"deploymentMode,omitempty"`
	// Runtime is the framework/language used to start with a linux's container an application.
	// It corresponds to one of the following values: spring-boot, vertx, thorntail, nodejs, python, php, ruby
	// It will be used to select the appropriate runtime image and logic
	Runtime string `json:"runtime,omitempty"`
	// Runtime's version
	Version string `json:"version,omitempty"`
	// To indicate if we want to expose the service out side of the cluster as a route
	ExposeService bool `json:"exposeService,omitempty"`
	// Port is the HTTP/TCP port number used within the pod by the runtime
	Port int32 `json:"port,omitempty"`
	// Storage allows to specify the capacity and mode of the volume to be mounted for the pod
	Storage Storage `json:"storage,omitempty"`
	// Array of env variables containing extra/additional info to be used to configure the runtime
	Envs []Env `json:"envs,omitempty"`
}

func (c ComponentSpec) GetImageName() string {
	return fmt.Sprintf("dev-runtime-%s", strings.ToLower(c.Runtime))
}
func (c ComponentSpec) GetImageReference() string {
	// todo: compute image version (and potentially name) based on runtime / version
	v := "latest"
	return util.GetImageReference(c.GetImageName(), v)
}

// ComponentStatus defines the observed state of Component
// +k8s:openapi-gen=true
type ComponentStatus struct {
	Phase     Phase `json:"phase,omitempty"`
	RevNumber string
	PodName   string       `json:"podName"`
	PodStatus v1.PodStatus `json:"podStatus"`
}

type Feature struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Id          int    `json:"id,omitempty"`
}

type Image struct {
	Name           string
	AnnotationCmds bool
	Repo           string
	Tag            string
	DockerImage    bool
}

type Storage struct {
	Name     string `json:"name,omitempty"`
	Capacity string `json:"capacity,omitempty"`
	Mode     string `json:"mode,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Component is the Schema for the components API
// +k8s:openapi-gen=true
type Component struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ComponentSpec   `json:"spec,omitempty"`
	Status ComponentStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ComponentList contains a list of Component
type ComponentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Component `json:"items"`
}
