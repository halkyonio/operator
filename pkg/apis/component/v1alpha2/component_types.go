package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DeploymentMode string

func (dm DeploymentMode) String() string {
	return string(dm)
}

const (
	DevDeploymentMode   DeploymentMode = "dev"
	BuildDeploymentMode DeploymentMode = "build"
)

// ComponentSpec defines the desired state of Component
// +k8s:openapi-gen=true
type ComponentSpec struct {
	// DeploymentMode indicates the strategy to be adopted to install the resources into a namespace
	// and next to create a pod. 2 strategies are currently supported; inner and outer loop
	// where outer loop refers to a build of the code and the packaging of the application into a container's image
	// while the inner loop will install a pod's running a supervisord daemon used to trigger actions such as : assemble, run, ...
	DeploymentMode DeploymentMode `json:"deploymentMode,omitempty"`
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
	Envs     []Env  `json:"envs,omitempty"`
	Revision string `json:"revision,omitempty"`
	// Build configuration used to execute a TekTon Build task
	BuildConfig BuildConfig `json:"buildConfig,omitempty"`
}

type ComponentPhase string

func (cp ComponentPhase) String() string {
	return string(cp)
}

const (
	// ComponentPending means the component has been accepted by the system, but it is still being processed. This includes time
	// before being bound to a node, as well as time spent pulling images onto the host, building and wiring capabilities.
	ComponentPending ComponentPhase = "Pending"
	// ComponentReady means the component is ready to accept code pushes
	ComponentReady ComponentPhase = "Ready"
	// ComponentRunning means the component has been bound to a node and all of its dependencies are available. The component is
	// able to process requests.
	ComponentRunning ComponentPhase = "Running"
	// ComponentSucceeded means that the component and its dependencies ran to successful completion
	// with a container exit code of 0, and the system is not going to restart any of these containers.
	ComponentSucceeded ComponentPhase = "Succeeded"
	// ComponentFailed means that the component and its dependencies have terminated, and at least one container has
	// terminated in a failure (exited with a non-zero exit code or was stopped by the system).
	ComponentFailed ComponentPhase = "Failed"
	// ComponentUnknown means that for some reason the state of the component could not be obtained, typically due
	// to an error in communicating with the host of the component.
	ComponentUnknown ComponentPhase = "Unknown"
	// ComponentBuilding means that the Build mode has been configured and that a build task is running
	ComponentBuilding ComponentPhase = "Building"
)

// ComponentStatus defines the observed state of Component
// +k8s:openapi-gen=true
type ComponentStatus struct {
	Phase   ComponentPhase `json:"phase,omitempty"`
	PodName string         `json:"podName,omitempty"`
	Message string         `json:"message,omitempty"`
}

type Storage struct {
	Name     string `json:"name,omitempty"`
	Capacity string `json:"capacity,omitempty"`
	Mode     string `json:"mode,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Component is the Schema for the components API
// +k8s:openapi-gen=true
// +genclient
type Component struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec    ComponentSpec   `json:"spec,omitempty"`
	Status  ComponentStatus `json:"status,omitempty"`
	requeue bool
}

func (in *Component) SetNeedsRequeue(requeue bool) {
	in.requeue = in.requeue || requeue
}

func (in *Component) NeedsRequeue() bool {
	return in.requeue
}

func (in *Component) isPending() bool {
	status := ComponentPending
	if BuildDeploymentMode == in.Spec.DeploymentMode {
		status = ComponentBuilding
	}
	return status == in.Status.Phase
}

func (in *Component) SetInitialStatus(msg string) bool {
	if !in.isPending() || in.Status.Message != msg {
		in.Status.Phase = ComponentPending
		if BuildDeploymentMode == in.Spec.DeploymentMode {
			in.Status.Phase = ComponentBuilding
		}
		in.Status.Message = msg
		return true
	}
	return false
}

func (in *Component) IsValid() bool {
	return true // todo: implement me
}

func (in *Component) SetErrorStatus(err error) bool {
	errMsg := err.Error()
	if ComponentFailed != in.Status.Phase || errMsg != in.Status.Message {
		in.Status.Phase = ComponentFailed
		in.Status.Message = errMsg
		return true
	}
	return false
}

func (in *Component) SetSuccessStatus(dependentName, msg string) bool {
	if dependentName != in.Status.PodName || ComponentReady != in.Status.Phase || msg != in.Status.Message {
		in.Status.Phase = ComponentReady
		in.Status.PodName = dependentName
		in.Status.Message = msg
		return true
	}
	return false
}

func (in *Component) GetStatusAsString() string {
	return in.Status.Phase.String()
}

func (in *Component) ShouldDelete() bool {
	return !in.DeletionTimestamp.IsZero()
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ComponentList contains a list of Component
type ComponentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Component `json:"items"`
}
