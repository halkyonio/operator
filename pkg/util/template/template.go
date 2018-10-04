/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package template

import (
	"bytes"
	"fmt"
	"github.com/shurcooL/httpfs/vfsutil"
	log "github.com/sirupsen/logrus"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha1"
	"os"
	"strings"
	"text/template"
)

var (
	TemplateAssets = Assets
	TemplatePath   = "innerloop"
	TemplateFiles  []string
	Templates      = make(map[string]template.Template)
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

		log.Debugf("Path of the file to be added as template : %s" + path)
		TemplateFiles = append(TemplateFiles, path)
		return nil
	}

	errW := vfsutil.Walk(TemplateAssets, TemplatePath, walkFn)
	if errW != nil {
		panic(errW)
	}

	// Fill an array with the k8s/openshift yaml files
	for i := range TemplateFiles {
		log.Debugf(" File : %s", TemplateFiles[i])

		// Create a new Template using the File name as key and add it to the array
		t := template.New(TemplateFiles[i])

		// Read Template's content
		data, err := vfsutil.ReadFile(TemplateAssets, TemplateFiles[i])
		if err != nil {
			log.Error(err)
		}
		t, err = t.Parse(bytes.NewBuffer(data).String())
		if err != nil {
			log.Error(err)
		}
		Templates[TemplateFiles[i]] = *t
	}
}

// Parse the file's template using the Component and the path of the template asset to use
func ParseTemplate(tmpl string, obj *v1alpha1.Component) bytes.Buffer {
	t := Templates[tmpl]
	return Parse(t, obj)
}

// Parse the file's template using the Application struct
func Parse(t template.Template, obj *v1alpha1.Component) bytes.Buffer {
	var b bytes.Buffer
	err := t.Execute(&b, obj)
	if err != nil {
		fmt.Println("There was an error:", err.Error())
	}
	log.Debug("Generated :", b.String())
	return b
}

//
func GetTemplateFullName(artifact string) string {
	return strings.Join([]string{"java", artifact}, "/")
}
