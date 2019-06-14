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

type BuildConfig struct {
	// URL is the Http or Web address of the Git repo to be cloned from the platforms such as : github, bitbucket, gitlab, ...
	// The syntax of the URL is : HTTPS://<git_server>/<git_org>/project.git
	URL string `json:"url"`
	// Ref is the git reference of the repo.
	// Optional. By default it is equal to "master".
	Ref string `json:"ref"`
	// ContextDir is a path to a subfolder in the repo. It could correspond to a maven module.
	// Optional. By default, it is equal to "."
	ContextDir string `json:"contextDir,omitempty"`
}
