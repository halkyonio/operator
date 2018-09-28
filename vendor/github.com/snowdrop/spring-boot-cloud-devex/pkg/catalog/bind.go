package catalog

import (
	appsv1 "github.com/openshift/api/apps/v1"
	appsocpv1 "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	corev1 "k8s.io/api/core/v1"

	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	servicecatalogclienset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"

	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"
	"github.com/snowdrop/spring-boot-cloud-devex/pkg/buildpack/types"
	"github.com/snowdrop/spring-boot-cloud-devex/pkg/common/oc"
	"k8s.io/apimachinery/pkg/util/uuid"
)

func Bind(config *restclient.Config, application types.Application, instance string, secret string) {
	serviceCatalogClient := GetClient(config)

	log.Infof("Let's generate a secret containing the parameters to be used by the application")
	createSecret(serviceCatalogClient, application.Namespace, instance, secret, nil, nil)
}

// Create a secret within the namespace of the service instance created using the service's parameters.
func createSecret(scc *servicecatalogclienset.ServicecatalogV1beta1Client, namespace, instanceName, secretName string,
	params interface{}, secrets map[string]string) error {

	// Generate UUID otherwise the binding's creation will fail if we use the same id as the instanceName, bindingName
	// todo: check I think this is needed since external id is optional
	UUID := string(uuid.NewUUID())

	_, err := scc.ServiceBindings(namespace).Create(
		&scv1beta1.ServiceBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      instanceName,
				Namespace: namespace,
			},
			Spec: scv1beta1.ServiceBindingSpec{
				ExternalID: UUID,
				ServiceInstanceRef: scv1beta1.LocalObjectReference{
					Name: instanceName,
				},
				SecretName:     secretName,
				Parameters:     BuildParameters(params),
				ParametersFrom: BuildParametersFrom(secrets),
			},
		})

	if err != nil {
		return errors.Wrap(err, "binding is failing")
	}
	log.Infof("Binding instance created")
	return nil
}

//Mount the secret as EnvFrom to the DeploymentConfig of the Application
func MountSecretAsEnvFrom(config *restclient.Config, application types.Application, secretName string) error {

	// Retrieve the DeploymentConfig
	deploymentConfigV1client := getAppsClient(config)
	deploymentConfigs := deploymentConfigV1client.DeploymentConfigs(application.Namespace)

	var dc *appsv1.DeploymentConfig
	var err error
	if oc.Exists("dc", application.Name) {
		dc, err = deploymentConfigs.Get(application.Name, metav1.GetOptions{})
		log.Infof("'%s' DeploymentConfig exists, got it", application.Name)
	}
	if err != nil {
		log.Fatalf("DeploymentConfig does not exist : %s", err.Error())
	}

	// Add the Secret as EnvVar to the container
	dc.Spec.Template.Spec.Containers[0].EnvFrom = addSecretAsEnvFromSource(secretName)

	// Update the DeploymentConfig
	_, errUpdate := deploymentConfigs.Update(dc)
	if errUpdate != nil {
		log.Fatalf("DeploymentConfig not updated : %s", errUpdate.Error())
	}
	log.Infof("'%s' EnvFrom secret added to the DeploymentConfig", secretName)

	// Redeploy it
	request := &appsv1.DeploymentRequest{
		Name:   application.Name,
		Latest: true,
		Force:  true,
	}

	_, errRedeploy := deploymentConfigs.Instantiate(application.Name, request)
	if errRedeploy != nil {
		log.Fatalf("Redeployment of the DeploymentConfig failed %s", errRedeploy.Error())
	}
	log.Infof("'%s' Deployment Config rollout succeeded to latest", secretName)

	return nil
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

func getAppsClient(config *restclient.Config) *appsocpv1.AppsV1Client {
	deploymentConfigV1client, err := appsocpv1.NewForConfig(config)
	if err != nil {
		log.Fatalf("Can't get DeploymentConfig Clientset: %s", err.Error())
	}
	return deploymentConfigV1client
}
