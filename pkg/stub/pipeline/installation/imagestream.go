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

package installation

import (
	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/apis/meta/v1"

	imagev1 "github.com/openshift/api/image/v1"
	imageclientsetv1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	"github.com/snowdrop/spring-boot-operator/pkg/apis/springboot/v1alpha1"
	"github.com/snowdrop/spring-boot-operator/pkg/util/oc"
	"github.com/snowdrop/spring-boot-operator/pkg/util/template"
	restclient "k8s.io/client-go/rest"
	"strings"
)

type Application struct {
	Name            string
	Version         string
	Namespace       string
	Replica         int
	Cpu             string `default:"100m"`
	Memory          string `default:"250Mi"`
	Port            int32  `default:"8080"`
	SupervisordName string
	Image           Image
}

type Image struct {
	Name           string
	AnnotationCmds bool
	Repo           string
	Tag            string
	DockerImage    bool
}

var (
	zero             = int64(0)
	deleteOptions    = &v1.DeleteOptions{GracePeriodSeconds: &zero}
	javaImage        = "quay.io/snowdrop/spring-boot-s2i"
	supervisordImage = "quay.io/snowdrop/supervisord"
)

// NewImageStreamStep creates a step that handles the creation of the DeploymentConfig
func NewImageStreamStep() Step {
	return &imageStreamStep{}
}

type imageStreamStep struct {
}

func (imageStreamStep) Name() string {
	return "imagestream"
}

func (imageStreamStep) CanHandle(springboot *v1alpha1.SpringBoot) bool {
	return true
}

func (imageStreamStep) Handle(springboot *v1alpha1.SpringBoot) error {
	// TODO
	return nil
}

var defaultImages = []Image{
	*CreateTypeImage(true, "dev-s2i", "latest", javaImage, false),
	*CreateTypeImage(true, "copy-supervisord", "latest", supervisordImage, true),
}

func CreateDefaultImageStreams(config *restclient.Config, appConfig Application) {
	CreateImageStreamTemplate(config, appConfig, defaultImages)
}

func CreateImageStreamTemplate(config *restclient.Config, appConfig Application, images []Image) {
	imageClient := getImageClient(config)
	appCfg := appConfig

	for _, img := range images {
		appCfg.Image = img

		// first check that the image stream hasn't already been created
		if oc.Exists("imagestream", img.Name) {
			log.Infof("'%s' ImageStream already exists, skipping", img.Name)
		} else {
			// Parse ImageStream Template
			tName := strings.Join([]string{template.BuilderPath, "imagestream"}, "/")
			var b = template.ParseTemplate(tName, appCfg)

			// Create ImageStream struct using the generated ImageStream string
			img := imagev1.ImageStream{}
			errYamlParsing := yaml.Unmarshal(b.Bytes(), &img)
			if errYamlParsing != nil {
				panic(errYamlParsing)
			}

			_, errImages := imageClient.ImageStreams(appConfig.Namespace).Create(&img)
			if errImages != nil {
				log.Fatalf("Unable to create ImageStream: %s", errImages.Error())
			}
		}
	}
}

func getImageClient(config *restclient.Config) *imageclientsetv1.ImageV1Client {
	imageClient, err := imageclientsetv1.NewForConfig(config)
	if err != nil {
		log.Fatal("Couldn't get ImageV1Client: %s", err)
	}
	return imageClient
}

func DeleteDefaultImageStreams(config *restclient.Config, appConfig Application) {
	for _, img := range defaultImages {
		// first check that the image stream hasn't already been created
		if oc.Exists("imagestream", img.Name) {
			client := getImageClient(config)
			err := client.ImageStreams(appConfig.Namespace).Delete(img.Name, deleteOptions)
			if err != nil {
				log.Fatalf("Unable to delete ImageStream: %s", img.Name)
			}
		}
	}
}

func CreateTypeImage(dockerImage bool, name string, tag string, repo string, annotationCmd bool) *Image {
	return &Image{
		DockerImage:    dockerImage,
		Name:           name,
		Repo:           repo,
		AnnotationCmds: annotationCmd,
		Tag:            tag,
	}
}
