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
	// moduleDirName is the name of the maven module / directory where build should be done.
	// Optional. By default, it is equal to "."
	ModuleDirName string `json:"moduleDirName,omitempty"`
	// contextPath is the directory relative to the git repository where the s2i build must take place.
	// Optional. By default, it is equal to "."
	ContextPath string `json:"contextPath,omitempty"`
	// Container image to be used as Base or From to build the final image
	BaseImage string `json:baseImage,omitempty`
}
