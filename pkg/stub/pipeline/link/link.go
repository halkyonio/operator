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
	v1 "github.com/openshift/api/apps/v1"
	appsocpv1 "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	log "github.com/sirupsen/logrus"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha1"
	"github.com/snowdrop/component-operator/pkg/stub/pipeline"
	"github.com/snowdrop/component-operator/pkg/util/kubernetes"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	return component.Status.Phase == ""
}

func (linkStep) Handle(component *v1alpha1.Component, deleted bool) error {
	target := component.DeepCopy()
	return createLink(target)
}

func createLink(component *v1alpha1.Component) error {
	// Get Current Namespace
	namespace, err := kubernetes.GetClientCurrentNamespace("")
	if err != nil {
		return err
	}

	component.ObjectMeta.Namespace = namespace
	componentName := component.Spec.Link.TargetComponentName
	secretName := component.Spec.Link.Ref

	// Get DeploymentConfig to inject EnvFrom using Secret and restart it
	dc, err := GetDeploymentConfig(namespace, componentName)
	if err != nil {
		return err
	}

	// Add the Secret as EnvVar to the container
	dc.Spec.Template.Spec.Containers[0].EnvFrom = addSecretAsEnvFromSource(secretName)

	// TODO -> Find a way to wait till service is up and running before to do the rollout
	//duration := time.Duration(10) * time.Second
	//time.Sleep(duration)

	// Update the DeploymentConfig
	err = sdk.Update(dc)
	if err != nil {
		log.Fatalf("DeploymentConfig not updated : %s", err.Error())
	}
	log.Infof("'%s' EnvFrom secret added to the DeploymentConfig", secretName)

	// Create a DeploymentRequest and redeploy it
	deploymentConfigV1client := getAppsClient()
	deploymentConfigs := deploymentConfigV1client.DeploymentConfigs(namespace)

	// Redeploy it
	request := &appsv1.DeploymentRequest{
		Name:   componentName,
		Latest: true,
		Force:  true,
	}

	_, errRedeploy := deploymentConfigs.Instantiate(componentName, request)
	if errRedeploy != nil {
		log.Fatalf("Redeployment of the DeploymentConfig failed %s", errRedeploy.Error())
	}

	log.Infof("%s link added", componentName)
	component.Status.Phase = v1alpha1.PhaseServiceCreation

	err = sdk.Update(component)
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func getAppsClient() *appsocpv1.AppsV1Client {
	config := kubernetes.GetK8RestConfig()
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

func GetDeploymentConfig(namespace string, name string) (*v1.DeploymentConfig, error) {
	dc := v1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps.openshift.io/v1",
			Kind:       "DeploymentConfig",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}
	if err := sdk.Get(&dc); err != nil {
		return nil, err
	}
	return &dc, nil
}
