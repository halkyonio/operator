package catalog

import (
	"encoding/json"
	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	servicecatalogclienset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"sort"
)

func List(config *restclient.Config, matching string) {
	serviceCatalogClient := GetClient(config)
	classes, _ := GetClusterServiceClasses(serviceCatalogClient)
	log.Info("List of services")
	log.Info("================")

	var filtered []scv1beta1.ClusterServiceClass
	if len(matching) > 0 {
		filtered = make([]scv1beta1.ClusterServiceClass, 0)
		for _, value := range classes {
			if strings.Contains(value.Spec.CommonServiceClassSpec.ExternalName, matching) {
				filtered = append(filtered, value)
			}
		}
	} else {
		filtered = classes
	}

	sort.Slice(filtered[:], func(i, j int) bool {
		return filtered[i].Spec.ExternalName < filtered[j].Spec.ExternalName
	})

	printListResults(filtered)
}

func printListResults(classes []scv1beta1.ClusterServiceClass) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetRowLine(true)
	table.SetHeader([]string{"Name", "Description", "Long Description"})
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("*")
	table.SetColumnSeparator("‡")
	table.SetRowSeparator("-")
	for _, class := range classes {
		var meta map[string]interface{}
		json.Unmarshal(class.Spec.ExternalMetadata.Raw, &meta)
		longDescription := ""
		if val, ok := meta["longDescription"]; ok {
			longDescription = val.(string)
		}
		row := []string{class.Spec.ExternalName, class.Spec.Description, longDescription}
		table.Append(row)
	}
	table.Render()
}

// GetClusterServiceClasses queries the service service catalog to get available clusterServiceClasses
func GetClusterServiceClasses(scc *servicecatalogclienset.ServicecatalogV1beta1Client) ([]scv1beta1.ClusterServiceClass, error) {
	classList, err := scc.ClusterServiceClasses().List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to list cluster service classes")
	}
	itemsFromAutomationBroker := make([]scv1beta1.ClusterServiceClass, 0)
	for _, class := range classList.Items {
		// we are only interested in services from the automation-service-broker
		// this broker however could have various names depending on the setup
		if strings.Contains(class.Spec.ClusterServiceBrokerName, "ansible-service-broker") ||
			strings.Contains(class.Spec.ClusterServiceBrokerName, "openshift-automation-service-broker") ||
			strings.Contains(class.Spec.ClusterServiceBrokerName, "automation-broker") {

			itemsFromAutomationBroker = append(itemsFromAutomationBroker, class)
		}
	}
	return itemsFromAutomationBroker, nil
}
