package v1alpha2

type BuildConfig struct {
	// Type is the mode that we would like to use to perform a container image build.
	// Optional. By default it is equal to s2i.
	Type string `json:type,omitempty`
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
