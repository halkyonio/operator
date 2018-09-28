package catalog

import (
	restclient "k8s.io/client-go/rest"

	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	servicecatalogclienset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"

	log "github.com/sirupsen/logrus"

	"encoding/json"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
)

func GetClient(config *restclient.Config) *servicecatalogclienset.ServicecatalogV1beta1Client {
	serviceCatalogV1Client, err := servicecatalogclienset.NewForConfig(config)
	if err != nil {
		log.Fatal("error creating service catalog Clientset", err.Error())
	}
	return serviceCatalogV1Client
}

// BuildParameters converts a map of variable assignments to a byte encoded json document,
// which is what the ServiceCatalog API consumes.
func BuildParameters(params interface{}) *runtime.RawExtension {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		// This should never be hit because marshalling a map[string]string is pretty safe
		// I'd rather throw a panic then force handling of an error that I don't think is possible.
		panic(fmt.Errorf("unable to marshal the request parameters %v (%s)", params, err))
	}

	return &runtime.RawExtension{Raw: paramsJSON}
}

// BuildParametersFrom converts a map of secrets names to secret keys to the
// type consumed by the ServiceCatalog API.
func BuildParametersFrom(secrets map[string]string) []scv1beta1.ParametersFromSource {
	params := make([]scv1beta1.ParametersFromSource, 0, len(secrets))

	for secret, key := range secrets {
		param := scv1beta1.ParametersFromSource{
			SecretKeyRef: &scv1beta1.SecretKeyReference{
				Name: secret,
				Key:  key,
			},
		}

		params = append(params, param)
	}
	return params
}
