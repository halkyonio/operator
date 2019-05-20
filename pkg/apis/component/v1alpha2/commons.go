package v1alpha2

const (
	// ComponentKind --
	ComponentKind string = "Component"

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
