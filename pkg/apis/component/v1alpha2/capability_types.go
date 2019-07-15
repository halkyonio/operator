package v1alpha2

import (
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

	/*
		      category: <database>, <logging>,<metrics>
			  kind: postgres (if category is database)
			  version: <version of the DB or prometheus or ...> to be installed
			  secretName: <secret_name_to_be_created> // Is used by kubedb postgres and is optional as some capability provider does not need to create a secret
			  parameters:
			     // LIST OF PARAMETERS WILL BE MAPPED TO EACH CATEGORY !
			    - name: DB_USER // WILL BE USED TO CREATE THE DB SECRET
			       value: "admin"
			    - name: DB_PASSWORD  // WILL BE USED TO CREATE THE DB SECRET
			       value: "admin"
	*/
	Category   CapabilityCategory `json:"category"`
	Kind       CapabilityKind     `json:"kind"`
	Version    string             `json:"version"`
	Parameters []Parameter        `json:"parameters,omitempty"`
}

type CapabilityCategory string
type CapabilityKind string

const (
	DatabaseCategory CapabilityCategory = "Database"
	PostgresKind     CapabilityKind     = "Postgres"

	MetricCategory  CapabilityCategory = "Metric"
	LoggingCategory CapabilityCategory = "Logging"
)

type CapabilityPhase string

func (c CapabilityPhase) String() string {
	return string(c)
}

const (
	// CapabilityPending means the capability has been accepted by the system, but it is still being processed. This includes time
	// being instantiated.
	CapabilityPending CapabilityPhase = "Pending"
	// CapabilityReady means the capability has been instantiated to a node and all of its dependencies are available. The
	// capability is able to process requests.
	CapabilityReady CapabilityPhase = "Ready"
	// CapabilityFailed means that the capability and its dependencies have terminated, and at least one container has
	// terminated in a failure (exited with a non-zero exit code or was stopped by the system).
	CapabilityFailed CapabilityPhase = "Failed"
)

// CapabilityStatus defines the observed state of Capability
// +k8s:openapi-gen=true
type CapabilityStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Phase   CapabilityPhase `json:"phase,omitempty"`
	PodName string          `json:"podName,omitempty"`
	Message string          `json:"message,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Capability is the Schema for the Services API
// +k8s:openapi-gen=true
// +genclient
type Capability struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec    CapabilitySpec   `json:"spec,omitempty"`
	Status  CapabilityStatus `json:"status,omitempty"`
	requeue bool
}

func (in *Capability) SetNeedsRequeue(requeue bool) {
	in.requeue = in.requeue || requeue
}

func (in *Capability) NeedsRequeue() bool {
	return in.requeue
}

func (in *Capability) SetInitialStatus(msg string) bool {
	if CapabilityPending != in.Status.Phase || in.Status.Message != msg {
		in.Status.Phase = CapabilityPending
		in.Status.Message = msg
		return true
	}
	return false
}

func (in *Capability) IsValid() bool {
	return true // todo: implement me
}

func (in *Capability) SetErrorStatus(err error) bool {
	errMsg := err.Error()
	if CapabilityFailed != in.Status.Phase || errMsg != in.Status.Message {
		in.Status.Phase = CapabilityFailed
		in.Status.Message = errMsg
		return true
	}
	return false
}

func (in *Capability) SetSuccessStatus(dependentName, msg string) bool {
	if dependentName != in.Status.PodName || CapabilityReady != in.Status.Phase || msg != in.Status.Message {
		in.Status.Phase = CapabilityReady
		in.Status.PodName = dependentName
		in.Status.Message = msg
		return true
	}
	return false
}

func (in *Capability) GetStatusAsString() string {
	return in.Status.Phase.String()
}

func (in *Capability) ShouldDelete() bool {
	return !in.DeletionTimestamp.IsZero()
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CapabilityList contains a list of Capability
type CapabilityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Capability `json:"items"`
}
