package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type LinkKind string

const (
	SecretLinkKind LinkKind = "Secret"
	EnvLinkKind    LinkKind = "Env"
)

func (l LinkKind) String() string {
	return string(l)
}

// LinkSpec defines the desired state of Link
// +k8s:openapi-gen=true
type LinkSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	ComponentName string   `json:"componentName"`
	Kind          LinkKind `json:"kind,omitempty"`
	Ref           string   `json:"ref,omitempty"`
	// Array of env variables containing extra/additional info to be used to configure the runtime
	Envs []Env `json:"envs,omitempty"`
}

type LinkPhase string

func (l LinkPhase) String() string {
	return string(l)
}

const (
	// LinkPending means the link has been accepted by the system, but it is still being processed.
	LinkPending LinkPhase = "Pending"
	// LinkReady means the link is ready.
	LinkReady LinkPhase = "Ready"
	// LinkFailed means that the linking operation failed.
	LinkFailed LinkPhase = "Failed"
	// LinkUnknown means that for some reason the state of the link could not be obtained, typically due
	// to an error in communicating with the host of the link.
	LinkUnknown LinkPhase = "Unknown"
)

// LinkStatus defines the observed state of Link
// +k8s:openapi-gen=true
type LinkStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Phase   LinkPhase `json:"phase,omitempty"`
	Message string    `json:"message"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Link is the Schema for the links API
// +k8s:openapi-gen=true
// +genclient
type Link struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec    LinkSpec   `json:"spec,omitempty"`
	Status  LinkStatus `json:"status,omitempty"`
	requeue bool
}

func (in *Link) SetNeedsRequeue(requeue bool) {
	in.requeue = in.requeue || requeue
}

func (in *Link) NeedsRequeue() bool {
	return in.requeue
}

func (in *Link) SetInitialStatus(msg string) bool {
	if LinkPending != in.Status.Phase || msg != in.Status.Message {
		in.Status.Phase = LinkPending
		in.Status.Message = msg
		return true
	}
	return false
}

func (in *Link) IsValid() bool {
	return true // todo: implement me
}

func (in *Link) SetErrorStatus(err error) bool {
	errMsg := err.Error()
	if LinkFailed != in.Status.Phase || errMsg != in.Status.Message {
		in.Status.Phase = LinkFailed
		in.Status.Message = errMsg
		return true
	}
	return false
}

func (in *Link) SetSuccessStatus(dependentName, msg string) bool {
	if LinkReady != in.Status.Phase || msg != in.Status.Message {
		in.Status.Phase = LinkReady
		in.Status.Message = msg
		return true
	}
	return false
}

func (in *Link) GetStatusAsString() string {
	return in.Status.Phase.String()
}

func (in *Link) ShouldDelete() bool {
	return !in.DeletionTimestamp.IsZero()
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LinkList contains a list of Link
type LinkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Link `json:"items"`
}
