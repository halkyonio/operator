package v1alpha2

type Phase string

const (
	// ComponentKind --
	ComponentKind string = "Component"

	// Phase Component --
	PhaseComponentCreation Phase = "CreatingComponent"
	PhaseComponentReady Phase = "Ready"
	PhaseComponentBuilding Phase = "Building"

	// Phase Linking --
	PhaseLinkCreation Phase = "CreatingLink"
	PhaseLinkReady    Phase = "Ready"

	// Phase Service --
	PhaseServiceCreation Phase = "CreatingService"
	PhaseServiceReady Phase = "Ready"

	// PhaseError --
	PhaseError Phase = "Error"

	// See https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/
	NameLabelKey           = "app.kubernetes.io/name"
	VersionLabelKey        = "app.kubernetes.io/version"
	InstanceLabelKey       = "app.kubernetes.io/instance"
	PartOfLabelKey         = "app.kubernetes.io/part-of"
	ComponentLabelKey      = "app.kubernetes.io/component"
	ManagedByLabelKey      = "app.kubernetes.io/managed-by"
	RuntimeLabelKey        = "app.openshift.io/runtime"
	RuntimeVersionLabelKey = "app.openshift.io/version"
)

type Env struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

type Parameter struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}
