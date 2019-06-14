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

package component

import (
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	pipeline "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

func (r *ReconcileComponent) installBuildMode(component *v1alpha2.Component, namespace string) (bool, error) {
	// Create Task s2i Buildah Push if it does not exists
	hasChanges := newFalse()
	if e := r.createAndCheckForChanges(component, &pipeline.Task{}, hasChanges); e != nil {
		return false, e
	}

	// Create ServiceAccount used by the Task's pod if it does not exists
	if e := r.createAndCheckForChanges(component, &v1.ServiceAccount{}, hasChanges); e != nil {
		return false, e
	}

	// TODO: oChange the status to mention that Build will start

	// Create the TaskRun in order to trigger the build
	if e := r.createAndCheckForChanges(component, &pipeline.TaskRun{}, hasChanges); e != nil {
		return false, e
	}

	return *hasChanges, nil
}

/*
func installOuterLoop(component v1alpha2.Component, config rest.Config, c client.Client, namespace string, scheme runtime.Scheme) error {
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

func fetchBuildConfig(c client.Client, component *v1alpha2.Component) (*build.BuildConfig, error) {
	log.Info("## Checking if the BuilConfig already exists")
	buildConfig := &build.BuildConfig{}
	err := c.Get(context.TODO(), types.NamespacedName{Name: component.Name, Namespace: component.Namespace}, buildConfig)
	return buildConfig, err
}

func fetchImageStream(c client.Client, component *v1alpha2.Component) (*image.ImageStream, error) {
	log.Info("## Checking if the ImageStream already exists")
	is := &image.ImageStream{}
	err := c.Get(context.TODO(), types.NamespacedName{Name: component.Name, Namespace: component.Namespace}, is)
	return is, err
}

func cloneDeploymentLoop(component v1alpha2.Component, config rest.Config, c client.Client, namespace string, scheme runtime.Scheme) error {
	component.ObjectMeta.Namespace = namespace

	isOpenshift, err := kubernetes.DetectOpenShift(&config)
	if err != nil {
		return err
	}

	if isOpenshift {
		tmpl, ok := util.Templates["outerloop/deploymentconfig"]
		if ok {
			originalcomponentName := component.Name

			// Populate the DC using template
			component.Name = component.Name + "-build"
			r, err := ParseTemplateToRuntimeObject(tmpl,&component)
			obj, err := kubernetes.RuntimeObjectFromUnstructured(r)
			if err != nil {
				return err
			}

			// Fetch DC
			dc := obj.(*deploymentconfig.DeploymentConfig)
			found, err := openshift.GetDeploymentConfig(namespace, originalcomponentName, c)
			if err != nil {
				log.Info("### DeploymentConfig not found")
				return err
			}
			containerFound := &found.Spec.Template.Spec.Containers[0]
			container := &dc.Spec.Template.Spec.Containers[0]
			container.Env = containerFound.Env
			container.EnvFrom = containerFound.EnvFrom
			container.Env = UpdateEnv(container.Env, component.Annotations["app.openshift.io/java-app-jar"])
			dc.Namespace = found.Namespace
			controllerutil.SetControllerReference(&component, dc, &scheme)

			err = c.Create(context.TODO(),dc)
			if err != nil {
				log.Info("### DeploymentConfig build creation failed")
				return err
			}
			log.Infof("### Created Build Deployment Config.")
		}
	}
	log.Info("## Pipeline 'outerloop' ended ##")
	log.Info("------------------------------------------------------")
	return nil
}

func UpdateEnv(envs []v1.EnvVar, jarName string) []v1.EnvVar {
	newEnvs := []v1.EnvVar{}
	for _, s := range envs {
		if s.Name == "JAVA_APP_JAR" {
			newEnvs = append(newEnvs, v1.EnvVar{Name: s.Name, Value: jarName})
		} else {
			newEnvs = append(newEnvs, s)
		}
	}
	return newEnvs
}

func updateSelector(component v1alpha2.Component, config rest.Config, c client.Client, namespace string, scheme runtime.Scheme) error {
	component.ObjectMeta.Namespace = namespace
	componentName := component.Annotations["app.openshift.io/component-name"]
	svc := &v1.Capability{}
	svc.Labels = map[string]string{
		"app.kubernetes.io/name": componentName,
		"app.openshift.io/runtime": component.Spec.Runtime,
	}
	if err := c.Get(context.TODO(), types.NamespacedName{Name: componentName, Namespace: namespace}, svc); err != nil {
		return err
	}

	var nameApp string
	if component.Spec.DeploymentMode == "outerloop" {
		nameApp = componentName + "-build"
	} else {
		nameApp = componentName
	}
	svc.Spec.Selector = map[string]string{
		"app": nameApp,
	}
	if err := c.Update(context.TODO(),svc) ; err != nil {
		return err
	}
	log.Infof("### Updated Capability Selector to switch to a different component.")
	log.Info("------------------------------------------------------------------")
	return nil
}
*/
