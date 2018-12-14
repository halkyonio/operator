package v1alpha1

// ComponentSpec defines the desired state of Component
type ComponentSpec struct {
	// Description represents a human readable string describing from a business perspective what this component is related to
	Description string
	// DeploymentMode indicates the strategy to be adopted to install the resources into a namespace
	// and next to create a pod. 2 strategies are currently supported; inner and outer loop
	// where outer loop refers to a build of the code and the packaging of the application into a container's image
	// while the inner loop will install a pod's running a supervisord daemon used to trigger actions such as : assemble, run, ...
	DeploymentMode string `json:"deploymentMode,omitempty"`
	// Runtime is the framework/language used to start with a linux's container an application.
	// It corresponds to one of the following values: spring-boot, vertx, thorntail, nodejs, python, php, ruby
	Runtime string `json:"runtime,omitempty"`
	// Runtime's version
	Version string `json:"version,omitempty"`
	// To indicate if we want to expose the service out side of the cluster as a route
	ExposeService bool `json:"exposeService,omitempty"`
	// Debug port is the port exposed for remote debugging
	DebugPort int `json:"debugPort,omitempty"`
	// Cpu is the cpu to be assigned to the pod running the application
	Cpu string `json:"cpu,omitempty"`
	// Cpu is the memory to be assigned to the pod running the application
	Memory string `json:"memory,omitempty"`
	// Port is the HTTP/TCP port number used within the pod by the runtime
	Port int32 `json:"ports,omitempty"`
	// The storage allows to specify the capacity and mode of the volume to be mounted for the pod
	Storage Storage `json:"storage,omitempty"`
	// Array of env variables containing extra/additional info to be used to configure the runtime
	Envs []Env `json:"envs,omitempty"`
	// List of services consumed by the runtime and created as service instance from a Service Catalog
	// and next binded to a DeploymentConfig using a service binding
	Services []Service `json:"services,omitempty"`
	// List of the links to other components
	Links []Link `json:"links,omitempty"`
}

// ComponentStatus defines the observed state of Component
type ComponentStatus struct {
	Phase Phase `json:"phase,omitempty"`
}

type Link struct {
	Name                string `json:"name,omitempty"`
	TargetComponentName string `json:"targetComponentName,omitempty"`
	Kind                string `json:"kind,omitempty"`
	Ref                 string `json:"ref,omitempty"`
	// Array of env variables containing extra/additional info to be used to configure the runtime
	Envs []Env `json:"envs,omitempty"`
}

type Env struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

type Service struct {
	Class          string      `json:"class,omitempty"`
	Name           string      `json:"name,omitempty"`
	Plan           string      `json:"plan,omitempty"`
	ExternalId     string      `json:"externalid,omitempty"`
	SecretName     string      `json:"secretname,omitempty"`
	Parameters     []Parameter `json:"parameters,omitempty"`
	ParametersJSon string
}

type Parameter struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

type Storage struct {
	Name     string `json:"name,omitempty"`
	Capacity string `json:"capacity,omitempty"`
	Mode     string `json:"mode,omitempty"`
}

// IntegrationPhase --
type Phase int

func (p Phase) String() string {
	return phases[p]
}

var phases = []string{
	"Unknown",
	"CreatingService",
	"Building",
	"Deploying",
	"Linking",
	"Error",
}

const (
	Unknown Phase = iota
	ServiceCreation
	Building
	Deploying
	Linking
	Error
)
