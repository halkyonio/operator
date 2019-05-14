package v1alpha2

type Phase string

const (
	// ComponentKind --
	ComponentKind string = "Component"
	LinkKind string = "Link"

	// PhaseServiceCreation --
	PhaseServiceCreation Phase = "CreatingService"
	// PhaseBuilding --
	PhaseBuilding Phase = "Building"
	// PhaseDeploying --
	PhaseDeploying Phase = "Deploying"
	// PhaseDeployed --
	PhaseDeployed Phase = "Deployed"
	// PhaseReady --
	PhaseReady Phase = "Ready"
	// PhaseLinking --
	PhaseLinking Phase = "Linking"
	// PhaseLinking --
	PhaseLinked Phase = "Linked"
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
