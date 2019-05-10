package link

import (
	appsv1 "github.com/openshift/api/apps/v1"
	appsocpv1 "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func (r *ReconcileLink) addSecretAsEnvFromSource(secretName string) corev1.EnvFromSource {
	return corev1.EnvFromSource{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: secretName},
			},
	}
}

func (r *ReconcileLink) addKeyValueAsEnvVar(key, value string) corev1.EnvVar {
	return corev1.EnvVar{
		Name:  key,
		Value: value,
	}
}

func (r *ReconcileLink) rolloutDeploymentConfig(name, namespace string) error {
	// Create a DeploymentRequest and redeploy it
	// As the Controller client can't process k8s sub-resource, then a separate
	// k8s client is needed
	deploymentConfigV1client, err := getAppsClient()
	if err != nil {
		return err
	}
	deploymentConfigs := deploymentConfigV1client.DeploymentConfigs(namespace)

	// Redeploy it
	request := &appsv1.DeploymentRequest{
		Name:   name,
		Latest: true,
		Force:  true,
	}

	_, err = deploymentConfigs.Instantiate(name, request)
	if err != nil && k8serrors.IsConflict(err) {
		// Can't rollout deployment config. We requeue
		return err
	}
	return nil
}

func getAppsClient() (*appsocpv1.AppsV1Client, error) {
	config, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	deploymentConfigV1client, err := appsocpv1.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return deploymentConfigV1client, nil
}