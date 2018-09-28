package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SpringBootList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []SpringBoot `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SpringBoot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              SpringBootSpec   `json:"spec"`
	Status            SpringBootStatus `json:"status,omitempty"`
}

type SpringBootSpec struct {
	// Fill me
}
type SpringBootStatus struct {
	// Fill me
}

const (
	// SpringBootKind --
	SpringBootKind string = "Spring Boot"
)
