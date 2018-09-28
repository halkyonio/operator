package scaffold

type Project struct {
	GroupId     string
	ArtifactId  string
	Version     string
	PackageName string
	OutDir      string
	Template    string `yaml:"template"  json:"template"`

	SnowdropBomVersion string
	SpringBootVersion  string
	Modules            []string

	UrlService string
}

type Config struct {
	Templates []Template `yaml:"templates"    json:"templates"`
	Boms      []Bom      `yaml:"bomversions"  json:"bomversions"`
	Modules   []Module   `yaml:"modules"      json:"modules"`
}

type Template struct {
	Name        string `yaml:"name"                     json:"name"`
	Description string `yaml:"description"              json:"description"`
}

type Bom struct {
	Community string `yaml:"community" json:"community"`
	Snowdrop  string `yaml:"snowdrop"  json:"snowdrop"`
	Default   bool   `yaml:"default"  json:"default"`
}

type Module struct {
	Name         string       `yaml:"name"             json:"name"`
	Description  string       `yaml:"description"      json:"description"`
	Guide        string       `yaml:"guide_ref"        json:"guide_ref"`
	Dependencies []Dependency `yaml:"dependencies"     json:"dependencies"`
	tags         []string     `yaml:"tags"             json:"tags"`
}

type Dependency struct {
	GroupId    string `yaml:"groupid"           json:"groupid"`
	ArtifactId string `yaml:"artifactid"        json:"artifactid"`
	Scope      string `yaml:"scope"             json:"scope"`
	Version    string `yaml:"version,omitempty" json:"version,omitempty"`
}
