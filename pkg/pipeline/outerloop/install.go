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

package outerloop

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha1"
	build "github.com/openshift/api/build/v1"
	image "github.com/openshift/api/image/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/snowdrop/component-operator/pkg/pipeline"
	"github.com/snowdrop/component-operator/pkg/util/kubernetes"
	util "github.com/snowdrop/component-operator/pkg/util/template"

	"k8s.io/apimachinery/pkg/runtime"
)

// NewInstallStep creates a step that handles the creation of the BuilcConfig
func NewInstallStep() pipeline.Step {
	return &installStep{}
}

type installStep struct{}

func (installStep) Name() string {
	return "deploy buildconfig"
}

func (installStep) CanHandle(component *v1alpha1.Component) bool {
	// log.Infof("## Status to be checked : %s", component.Status.Phase)
	return true
}

func (installStep) Handle(component *v1alpha1.Component, config *rest.Config, client *client.Client, namespace string, scheme *runtime.Scheme) error {
	return installOuterLoop(*component, *config, *client, namespace, *scheme)
}

func installOuterLoop(component v1alpha1.Component, config rest.Config, c client.Client, namespace string, scheme runtime.Scheme) error {
	log.Info("Install BuildConfig ...")
	component.ObjectMeta.Namespace = namespace

	isOpenshift, err := kubernetes.DetectOpenShift(&config)
	if err != nil {
		return err
	}

	if isOpenshift {
		tmpl, ok := util.Templates["outerloop/imagestream"]
		if ok {
			// Check if an ImageStream already exists
			is, err := fetchImageStream(c, &component)
			if err != nil {
				err = kubernetes.CreateResource(tmpl, &component, c, &scheme)
				if err != nil {
					return err
				}
				log.Infof("### Created ImageStream used as target image to run the application")
			} else {
				log.Infof("### Image stream already exists %s",is.Name)
			}
		}

		tmpl, ok = util.Templates["outerloop/buildconfig"]
		if ok {
			// Check if a BuildConfig already exists
			bc, err := fetchBuildConfig(c, &component)
			if err != nil {
				err := kubernetes.CreateResource(tmpl, &component, c, &scheme)
				if err != nil {
					return err
				}
				log.Infof("### Created Buildconfig")
			} else {
				log.Infof("### BuildConfig already exists: %s",bc.Name)
			}
		}
	}
	return nil
}

func fetchBuildConfig(c client.Client, component *v1alpha1.Component) (*build.BuildConfig, error) {
	log.Info("## Checking if the BuilConfig already exists")
	buildConfig := &build.BuildConfig{}
	err := c.Get(context.TODO(), types.NamespacedName{Name: component.Name, Namespace: component.Namespace}, buildConfig)
	return buildConfig, err
}

func fetchImageStream(c client.Client, component *v1alpha1.Component) (*image.ImageStream, error) {
	log.Info("## Checking if the ImageStream already exists")
	is := &image.ImageStream{}
	err := c.Get(context.TODO(), types.NamespacedName{Name: component.Name, Namespace: component.Namespace}, is)
	return is, err
}
