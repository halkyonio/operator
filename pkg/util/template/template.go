package template

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"go/types"
	"text/template"
)

var (
	templates = make(map[string]template.Template)
)

const (
	BuilderPath = "java"
	PvcName     = "m2-data"
)

// Parse the file's template using the Application struct
func ParseTemplate(tmpl string, cfg types.Object) bytes.Buffer {
	// Create Template and parse it
	var b bytes.Buffer
	t := templates[tmpl]
	err := t.Execute(&b, cfg)
	if err != nil {
		fmt.Println("There was an error:", err.Error())
	}
	log.Debug("Generated :", b.String())
	return b
}
