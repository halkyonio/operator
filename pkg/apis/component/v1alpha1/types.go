package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ComponentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Component `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Component struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ComponentSpec   `json:"spec"`
	Status            ComponentStatus `json:"status,omitempty"`
}

type ComponentSpec struct {
	DeploymentMode  string `json:"deployment,omitempty"`
	Runtime         string `json:"runtime,omitempty"`
	ExposeService   bool   `json:"exposeService,omitempty"`
	Cpu             string `json:"cpu,omitempty"`
	Memory          string `json:"memory,omitempty"`
	Port            int32  `json:"port,omitempty"`
	SupervisordName string
	Storage         Storage `json:"storage,omitempty"`
	Images          []Image `json:"image,omitempty"`
	Envs            []Env   `json:"env,omitempty"`
	Services        []Service
}

type ComponentStatus struct {
	Phase Phase `json:"phase,omitempty"`
}

type Image struct {
	Name           string
	AnnotationCmds bool
	Repo           string
	Tag            string
	DockerImage    bool
}

type Env struct {
	Name  string
	Value string
}

type Service struct {
	Class      string
	Name       string
	Plan       string `default:"dev"`
	ExternalId string
	Parameters []Parameter
}

type Parameter struct {
	Name  string
	Value string
}

type Storage struct {
	Name     string `json:"name,omitempty"`
	Capacity string `json:"capacity,omitempty"`
	Mode     string `json:"mode,omitempty"`
}

// IntegrationPhase --
type Phase string

const (
	// ComponentKind --
	ComponentKind string = "Component"

	// PhaseBuilding --
	PhaseBuilding Phase = "Building"
	// PhaseDeploying --
	PhaseDeploying Phase = "Deploying"
	// PhaseRunning --
	PhaseRunning Phase = "Running"
	// PhaseError --
	PhaseError Phase = "Error"
)
