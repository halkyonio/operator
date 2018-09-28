package catalog

import (
	"github.com/snowdrop/spring-boot-cloud-devex/pkg/buildpack/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"

	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	servicecatalogclienset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"
)

func Create(config *restclient.Config, application types.Application, instanceName string) {
	serviceCatalogClient := GetClient(config)
	log.Infof("Service instance '%s' will be created ...", instanceName)

	service, e := application.GetService(instanceName)
	if e != nil {
		log.Errorf("No such service definition '%s' found in MANIFEST", instanceName)
	} else {
		createServiceInstance(serviceCatalogClient, application.Namespace, instanceName, service.Class, service.Plan, service.ExternalId, service.ParametersAsMap())
		log.Infof("Service instance created")
	}

}

// CreateServiceInstance creates service instance from service catalog
func createServiceInstance(scc *servicecatalogclienset.ServicecatalogV1beta1Client, ns string, instanceName string, className string, plan string, externalID string, params interface{}) error {

	// Creating Service Instance
	_, err := scc.ServiceInstances(ns).Create(
		&scv1beta1.ServiceInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      instanceName,
				Namespace: ns,
			},
			Spec: scv1beta1.ServiceInstanceSpec{
				ExternalID: externalID,
				PlanReference: scv1beta1.PlanReference{
					ClusterServiceClassExternalName: className,
					ClusterServicePlanExternalName:  plan,
				},
				Parameters: BuildParameters(params),
			},
		})

	if err != nil {
		return errors.Wrap(err, "unable to create service instance")
	}
	return nil
}
