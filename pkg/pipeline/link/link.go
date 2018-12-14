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
	"github.com/openshift/api/apps/v1"
	appsv1 "github.com/openshift/api/apps/v1"
	appsocpv1 "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	api "github.com/snowdrop/component-api/pkg/apis/component/v1alpha1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha1"
	"github.com/snowdrop/component-operator/pkg/pipeline"
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"time"
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

func (linkStep) CanHandle(component *v1alpha1.Component) bool {
	return component.Status.Phase == api.ServiceCreation || component.Status.Phase == api.Deploying || component.Status.Phase == api.Unknown
}

func (linkStep) Handle(component *v1alpha1.Component, client *client.Client, namespace string) error {
	return createLink(component, *client, namespace)
}

func createLink(component *v1alpha1.Component, c client.Client, namespace string) error {
	retryInterval, _ := time.ParseDuration("10s")
	component.ObjectMeta.Namespace = namespace
	for _, l := range component.Spec.Links {
		componentName := l.TargetComponentName
		if componentName != "" {
			// Get DeploymentConfig to inject EnvFrom using Secret and restart it
			err := wait.Poll(retryInterval, time.Duration(2)*retryInterval, func() (done bool, err error) {
				dc, err := GetDeploymentConfig(namespace, componentName, c)
				if err != nil {
					return false, err
				}
				logMessage := ""
				kind := l.Kind
				switch kind {
				case "Secret":
					secretName := l.Ref
					// Add the Secret as EnvVar to the container
					dc.Spec.Template.Spec.Containers[0].EnvFrom = addSecretAsEnvFromSource(secretName)
					logMessage = "#### Added the deploymentConfig's EnvFrom reference of the secret " + secretName
				case "Env":
					// TODO Iterate through Env vars
					key := l.Envs[0].Name
					val := l.Envs[0].Value
					dc.Spec.Template.Spec.Containers[0].Env = append(dc.Spec.Template.Spec.Containers[0].Env, addKeyValueAsEnvVar(key, val))
					logMessage = "#### Added the deploymentConfig's EnvVar : " + key + ", " + val
				}

				// Update the DeploymentConfig
				err = c.Update(context.TODO(), dc)
				if err != nil && k8serrors.IsConflict(err) {
					// Retry function on conflict
					return false, nil
				}
				if err != nil {
					return false, err
				}
				log.Info(logMessage)

				// Create a DeploymentRequest and redeploy it
				// As the Controller client can't process k8s sub-resource, then a separate
				// k8s client is needed
				deploymentConfigV1client := getAppsClient()
				deploymentConfigs := deploymentConfigV1client.DeploymentConfigs(namespace)

				// Redeploy it
				request := &appsv1.DeploymentRequest{
					Name:   componentName,
					Latest: true,
					Force:  true,
				}

				_, errRedeploy := deploymentConfigs.Instantiate(componentName, request)
				if errRedeploy != nil && k8serrors.IsConflict(err) {
					// Retry reconcile
					return false, nil
				}
				if err != nil {
					return false, err
				}
				log.Infof("#### Added %s link's CRD component", componentName)
				log.Infof("#### Rollout the DeploymentConfig of the '%s' component", component.Name)

				return true, nil
			})
			if err != nil {
				return err
			}
		} else {
			return errors.New("Target component is not defined !!")
		}
	}
	component.Status.Phase = api.Linking
	err := c.Update(context.TODO(), component)
	if err != nil && k8serrors.IsConflict(err) {
		return err
	}
	log.Info("### Pipeline 'link' ended ###")
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

func GetDeploymentConfig(namespace string, name string, c client.Client) (*v1.DeploymentConfig, error) {
	dc := &v1.DeploymentConfig{}
	if err := c.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, dc); err != nil {
		return nil, err
	}
	return dc, nil
}
