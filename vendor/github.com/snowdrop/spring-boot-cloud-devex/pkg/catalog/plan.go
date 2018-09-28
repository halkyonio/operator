package catalog

import (
	"encoding/json"
	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	servicecatalogclienset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"
	"os"
	"strconv"
)

type serviceInstanceCreateParameterSchema struct {
	Required   []string
	Properties map[string]property
}

type property struct {
	Title       string
	Type        string
	Description string
}

type propertyOut struct {
	Name        string
	Title       string
	Description string
	Type        string
	Required    bool
}

func Plan(config *restclient.Config, className string) {
	serviceCatalogClient := GetClient(config)
	matchingPlans, err := getMatchingPlans(serviceCatalogClient, className)
	if err != nil {
		log.Fatal(err)
	}

	results := make(map[string][]propertyOut)
	for _, plan := range matchingPlans {
		properties, err := getProperties(&plan)
		if err != nil {
			log.Fatal(err)
		}
		results[plan.Spec.ExternalName] = properties
	}
	printPlanResults(results)
}

func getMatchingPlans(scc *servicecatalogclienset.ServicecatalogV1beta1Client, className string) ([]scv1beta1.ClusterServicePlan, error) {
	clusterServiceClasses, err := GetClusterServiceClasses(scc)
	if err != nil {
		log.Fatalf("Unable to list ClusterServiceClasses")
		return nil, errors.Wrap(err, "Unable to list ClusterServiceClasses")
	}

	var matchingClusterServiceClass *scv1beta1.ClusterServiceClass
	for _, c := range clusterServiceClasses {
		if c.Spec.ExternalName == className {
			matchingClusterServiceClass = &c
			break
		}
	}
	if matchingClusterServiceClass == nil {
		log.Fatalf("Unable to locate ClusterServiceClasses with name '%s'", className)
		return nil, errors.Wrapf(err, "Unable to locate ClusterServiceClasses with name '%s'", className)
	}

	planList, err := scc.ClusterServicePlans().List(metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Unable to list ClusterServicePlans")
	}

	matchingPlans := make([]scv1beta1.ClusterServicePlan, 0)
	for _, p := range planList.Items {
		if p.Spec.ClusterServiceClassRef.Name == matchingClusterServiceClass.Spec.ExternalID {
			matchingPlans = append(matchingPlans, p)
		}
	}

	return matchingPlans, nil
}

func getProperties(plan *scv1beta1.ClusterServicePlan) ([]propertyOut, error) {
	paramBytes := plan.Spec.CommonServicePlanSpec.ServiceInstanceCreateParameterSchema.Raw
	schema := serviceInstanceCreateParameterSchema{}

	err := json.Unmarshal(paramBytes, &schema)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable unmashal response: %s", string(paramBytes[:]))
	}

	result := make([]propertyOut, 0)
	for k, v := range schema.Properties {
		propertyOut := propertyOut{}
		propertyOut.Name = k
		propertyOut.Title = v.Title
		propertyOut.Description = v.Description
		propertyOut.Type = v.Type
		propertyOut.Required = isRequired(schema.Required, k)
		result = append(result, propertyOut)
	}

	return result, nil
}

func isRequired(required []string, name string) bool {
	for _, n := range required {
		if n == name {
			return true
		}
	}
	return false
}

func printPlanResults(results map[string][]propertyOut) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetRowLine(true)
	table.SetHeader([]string{"Plan", "Property Name", "Required", "Type", "Description", "Long Description"})
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("*")
	table.SetColumnSeparator("‡")
	table.SetRowSeparator("-")

	for plan, properties := range results {
		for _, property := range properties {
			row := []string{plan, property.Name, strconv.FormatBool(property.Required), property.Type, property.Title, property.Description}
			table.Append(row)
		}
		table.Append([]string{"", "", "", ""})
	}

	table.Render()
}
