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
	Version         string
	Namespace       string
	Cpu             string `default:"100m"`
	Memory          string `default:"250Mi"`
	Port            int32  `default:"8080"`
	SupervisordName string
	Image           Image
}

type ComponentStatus struct {
	// Fill me
}

type Image struct {
	Name           string
	AnnotationCmds bool
	Repo           string
	Tag            string
	DockerImage    bool
}

const (
	// ComponentKind --
	ComponentKind string = "Component"
)
