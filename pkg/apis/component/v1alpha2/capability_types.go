package v1alpha2

import (
	servicecatalogv1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CapabilitySpec defines the desired state of Capability
// +k8s:openapi-gen=true
type CapabilitySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html

	// Info needed to create a Capability instance from the Capability Catalog
	// and next to bind the info to a Deployment using the Capability binding/secret
	Class          string      `json:"class,omitempty"`
	Name           string      `json:"name,omitempty"`
	Plan           string      `json:"plan,omitempty"`
	ExternalId     string      `json:"externalid,omitempty"`
	SecretName     string      `json:"secretname,omitempty"`
	Parameters     []Parameter `json:"parameters,omitempty"`
	ParametersJSon string
}

// CapabilityStatus defines the observed state of Capability
// +k8s:openapi-gen=true
type CapabilityStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Phase                 Phase `json:"phase,omitempty"`
	ServiceBindingName    string `json:"serviceBindingName"`
	ServiceBindingStatus  servicecatalogv1.ServiceBindingStatus`json:"serviceBindingStatus"`
	ServiceInstanceName   string `json:"serviceInstanceName"`
	ServiceInstanceStatus servicecatalogv1.ServiceInstanceStatus`json:"serviceInstanceStatus"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Capability is the Schema for the Services API
// +k8s:openapi-gen=true
type Capability struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CapabilitySpec   `json:"spec,omitempty"`
	Status CapabilityStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CapabilityList contains a list of Capability
type CapabilityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Capability `json:"items"`
}
