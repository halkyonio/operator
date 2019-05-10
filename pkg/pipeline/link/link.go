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

package link

import (
	appsv1 "github.com/openshift/api/apps/v1"
	appsocpv1 "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	//"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"github.com/snowdrop/component-operator/pkg/pipeline"
	"github.com/snowdrop/component-operator/pkg/util/kubernetes"
	//"github.com/snowdrop/component-operator/pkg/util/openshift"
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	//"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	//"time"
)

// NewLinkStep creates a step that handles the creation of the Service from the catalog
func NewLinkStep() pipeline.Step {
	return &linkStep{}
}

type linkStep struct {
}

func (linkStep) Name() string {
	return "link"
}

func (linkStep) CanHandle(component *v1alpha2.Component) bool {
	 // log.Infof("## Status to be checked : %s", component.Status.Phase)
	 return component.Status.Phase == v1alpha2.PhaseServiceCreation ||  component.Status.Phase == v1alpha2.PhaseDeploying || component.Status.Phase == ""
}

func (linkStep) Handle(component *v1alpha2.Component, config *rest.Config, client *client.Client, namespace string, scheme *runtime.Scheme) error {
	return createLink(*component, *config, *client, namespace, *scheme)
}

func createLink(component v1alpha2.Component, cfg rest.Config, c client.Client, namespace string, scheme runtime.Scheme) error {
	//_, _ := time.ParseDuration("10s")
	component.ObjectMeta.Namespace = namespace

	_, err := kubernetes.DetectOpenShift(&cfg)
	if err != nil {
		return err
	}

/*	for _, l := range component.Spec.Links {
		componentName := l.TargetComponentName
		if componentName != "" {
			// Get DeploymentConfig to inject EnvFrom using Secret and restart it
			err := wait.Poll(retryInterval, time.Duration(2)*retryInterval, func() (done bool, err error) {
				if (isOpenshift) {
					found, err := openshift.GetDeploymentConfig(namespace, componentName, c)
					if err != nil {
						log.Info("### DeploymentConfig not found")
						return false, err
					}
					logMessage := ""
					kind := l.Kind
					switch kind {
					case "Secret":
						secretName := l.Ref
						// Add the Secret as EnvVar to the container
						found.Spec.Template.Spec.Containers[0].EnvFrom = addSecretAsEnvFromSource(secretName)
						logMessage = "### Added the deploymentConfig's EnvFrom reference of the secret " + secretName
					case "Env":
						// TODO Iterate through Env vars
						key := l.Envs[0].Name
						val := l.Envs[0].Value
						found.Spec.Template.Spec.Containers[0].Env = append(found.Spec.Template.Spec.Containers[0].Env, addKeyValueAsEnvVar(key, val))
						logMessage = "### Added the deploymentConfig's EnvVar : " + key + ", " + val
					}

					// Update the DeploymentConfig
					err = c.Update(context.TODO(), found)
					if err != nil && k8serrors.IsConflict(err) {
						// Retry function on conflict
						log.Info("### DeploymentConfig update failed due to conflict!")
						return false, nil
					}
					if err != nil {
						log.Info("### DeploymentConfig update failed !")
						return false, err
					}
					log.Info("### DeploymentConfig updated")
					log.Info(logMessage)

					// Rollout the DC
					err = rolloutDeploymentConfig(componentName, namespace)
					if err != nil {
						log.Info("Deployment Config rollout failed !")
						return false, err
					}

					log.Infof("### Added %s link's CRD component", componentName)
					log.Infof("### Rollout the DeploymentConfig of the '%s' component", component.Name)
					return true, nil
				} else {
					// K8s platform. We will fetch a deployment
					d, err := kubernetes.GetDeployment(namespace, componentName, c)
					if err != nil {
						return false, err
					}
					logMessage := ""
					kind := l.Kind
					switch kind {
					case "Secret":
						secretName := l.Ref
						// Add the Secret as EnvVar to the container
						d.Spec.Template.Spec.Containers[0].EnvFrom = addSecretAsEnvFromSource(secretName)
						logMessage = "### Added the deploymentConfig's EnvFrom reference of the secret " + secretName
					case "Env":
						// TODO Iterate through Env vars
						key := l.Envs[0].Name
						val := l.Envs[0].Value
						d.Spec.Template.Spec.Containers[0].Env = append(d.Spec.Template.Spec.Containers[0].Env, addKeyValueAsEnvVar(key, val))
						logMessage = "### Added the deploymentConfig's EnvVar : " + key + ", " + val
					}

					// Update the Deployment
					err = c.Update(context.TODO(), d)
					if err != nil && k8serrors.IsConflict(err) {
						// Retry function on conflict
						return false, nil
					}
					if err != nil {
						return false, err
					}
					log.Info(logMessage)

					log.Infof("### Added %s link's CRD component", componentName)
					log.Infof("### Rollout Deployment of the '%s' component", component.Name)
					return true, nil
				}
			})
			if err != nil {
				return err
			}
		} else {
			return errors.New("Target component is not defined !!")
		}
	}*/

	log.Info("### Component Link updated.")

	component.Status.Phase = v1alpha2.PhaseLinking
	// err = c.Status().Update(context.TODO(), found)
	err = c.Update(context.TODO(),&component)
	if err != nil && k8serrors.IsNotFound(err) {
		log.Info("## Component link - status update failed")
		return err
	}
	log.Info("## Pipeline 'link' ended ##")
	log.Infof("## Status updated : %s ##",component.Status.Phase)
	log.Infof("## Status RevNumber : %s ##",component.Status.RevNumber)
	log.Info("------------------------------------------------------")
	return nil
}

func rolloutDeploymentConfig(name, namespace string) error {
	// Create a DeploymentRequest and redeploy it
	// As the Controller client can't process k8s sub-resource, then a separate
	// k8s client is needed
	deploymentConfigV1client := getAppsClient()
	deploymentConfigs := deploymentConfigV1client.DeploymentConfigs(namespace)

	// Redeploy it
	request := &appsv1.DeploymentRequest{
		Name:   name,
		Latest: true,
		Force:  true,
	}

	_, err := deploymentConfigs.Instantiate(name, request)
	if err != nil && k8serrors.IsConflict(err) {
		// Can't rollout deployment config. We requeue
		return err
	}
	return nil
}

func getAppsClient() *appsocpv1.AppsV1Client {
	config, err := config.GetConfig()
	if err != nil {
		log.Fatalf("Can't get the K8s config: %s", err.Error())
	}
	deploymentConfigV1client, err := appsocpv1.NewForConfig(config)
	if err != nil {
		log.Fatalf("Can't get DeploymentConfig Clientset: %s", err.Error())
	}
	return deploymentConfigV1client
}

func addSecretAsEnvFromSource(secretName string) []corev1.EnvFromSource {
	return []corev1.EnvFromSource{
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: secretName},
			},
		},
	}
}

func addKeyValueAsEnvVar(key, value string) corev1.EnvVar {
	return corev1.EnvVar{
		Name:  key,
		Value: value,
	}
}