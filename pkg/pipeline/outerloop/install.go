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
	log "github.com/sirupsen/logrus"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha1"
	"golang.org/x/net/context"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/snowdrop/component-operator/pkg/pipeline"
	"github.com/snowdrop/component-operator/pkg/util/kubernetes"
	util "github.com/snowdrop/component-operator/pkg/util/template"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"text/template"
)

// NewInstallStep creates a step that handles the creation of the BuilcConfig
func NewInstallStep() pipeline.Step {
	return &installStep{}
}

type installStep struct{}

func (installStep) Name() string {
	return "deploy"
}

func (installStep) CanHandle(component *v1alpha1.Component) bool {
	// log.Infof("## Status to be checked : %s", component.Status.Phase)
	return true
}

func (installStep) Handle(component *v1alpha1.Component, client *client.Client, namespace string, scheme *runtime.Scheme) error {
	return installOuterLoop(component, *client, namespace, *scheme)
}

func installOuterLoop(component *v1alpha1.Component, c client.Client, namespace string, scheme runtime.Scheme) error {
	log.Info("Install BuildConfig ...")
	component.ObjectMeta.Namespace = namespace

	isOpenshift, err := kubernetes.DetectOpenShift()
	if err != nil {
		return err
	}

	if isOpenshift {
		tmpl, ok := util.Templates["outerloop/imagestream"]
		if ok {
			err := createResource(tmpl, component, c, &scheme)
			if err != nil {
				return err
			}
			log.Infof("### Created Imagestream used as target image to run the application")
		}

		tmpl, ok = util.Templates["outerloop/buildconfig"]
		if ok {
			err := createResource(tmpl, component, c, &scheme)
			if err != nil {
				return err
			}
			log.Infof("### Created Buildconfig")
		}
	}

	log.Info("## Pipeline 'outerloop' ended ##")
	log.Info("------------------------------------------------------")
	return nil
}

func createResource(tmpl template.Template, component *v1alpha1.Component, c client.Client, scheme *runtime.Scheme) error {
	res, err := newResourceFromTemplate(tmpl, component, scheme)
	if err != nil {
		return err
	}

	for _, r := range res {
		if obj, ok := r.(metav1.Object); ok {
			obj.SetLabels(kubernetes.PopulateK8sLabels(component, "Backend"))
		}
		err = c.Create(context.TODO(), r)
		if err != nil && k8serrors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

func newResourceFromTemplate(template template.Template, component *v1alpha1.Component, scheme *runtime.Scheme) ([]runtime.Object, error) {
	var result = []runtime.Object{}

	var b = util.Parse(template, component)
	r, err := kubernetes.PopulateKubernetesObjectFromYaml(b.String())
	if err != nil {
		return nil, err
	}

	if strings.HasSuffix(r.GetKind(), "List") {
		l, err := r.ToList()
		if err != nil {
			return nil, err
		}
		for _, item := range l.Items {
			obj, err := kubernetes.RuntimeObjectFromUnstructured(&item)
			if err != nil {
				return nil, err
			}
			ro, ok := obj.(v1.Object)
			ro.SetNamespace(component.Namespace)
			if !ok {
				return nil, err
			}
			controllerutil.SetControllerReference(component, ro, scheme)
			//kubernetes.SetNamespaceAndOwnerReference(obj, component)
			result = append(result, obj)
		}
	} else {
		obj, err := kubernetes.RuntimeObjectFromUnstructured(r)
		if err != nil {
			return nil, err
		}

		ro, ok := obj.(v1.Object)
		ro.SetNamespace(component.Namespace)
		if !ok {
			return nil, err
		}
		controllerutil.SetControllerReference(component, ro, scheme)
		//kubernetes.SetNamespaceAndOwnerReference(obj, component)
		result = append(result, obj)
	}
	return result, nil
}
