package buildpack

import (
	"bytes"
	"fmt"
	"github.com/shurcooL/httpfs/vfsutil"
	log "github.com/sirupsen/logrus"
	"os"
	"text/template"

	"github.com/snowdrop/spring-boot-cloud-devex/pkg/buildpack/types"
)

var (
	assetsBuildPackTemplates = Assets
	templateBuildPackFiles   []string
	templates                = make(map[string]template.Template)
)

const (
	builderPath = "java"
)

func init() {
	walkFn := func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			log.Printf("can't stat file %s: %v\n", path, err)
			return nil
		}

		if fi.IsDir() {
			return nil
		}

		log.Debugf("Path of the buildpack file to be added as template : %s" + path)
		templateBuildPackFiles = append(templateBuildPackFiles, path)
		return nil
	}

	errW := vfsutil.Walk(assetsBuildPackTemplates, builderPath, walkFn)
	if errW != nil {
		panic(errW)
	}

	// Fill an array with our Builder's text/template
	for i := range templateBuildPackFiles {
		log.Debugf("BuildPack File : %s", templateBuildPackFiles[i])

		// Create a new Template using the File name as key and add it to the array
		t := template.New(templateBuildPackFiles[i])

		// Read Template's content
		data, err := vfsutil.ReadFile(assetsBuildPackTemplates, templateBuildPackFiles[i])
		if err != nil {
			log.Error(err)
		}
		t, err = t.Parse(bytes.NewBuffer(data).String())
		if err != nil {
			log.Error(err)
		}
		templates[templateBuildPackFiles[i]] = *t
	}
}

// Parse the file's template using the Application struct
func ParseTemplate(tmpl string, cfg types.Application) bytes.Buffer {
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
