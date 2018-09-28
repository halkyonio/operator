package buildpack

import (
	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"

	restclient "k8s.io/client-go/rest"

	routev1 "github.com/openshift/api/route/v1"
	routeclientsetv1 "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"

	"github.com/snowdrop/spring-boot-cloud-devex/pkg/buildpack/types"
	"github.com/snowdrop/spring-boot-cloud-devex/pkg/common/oc"
	"strings"
)

func CreateRouteTemplate(config *restclient.Config, application types.Application) {
	if oc.Exists("route", application.Name) {
		log.Infof("'%s' Route already exists, skipping", application.Name)
	} else {
		routeV1Client := getClient(config)

		// Parse Route Template
		tName := strings.Join([]string{builderPath, "route"}, "/")
		var b = ParseTemplate(tName, application)

		// Create Route struct using the generated Route string
		route := routev1.Route{}
		errYamlParsing := yaml.Unmarshal(b.Bytes(), &route)
		if errYamlParsing != nil {
			panic(errYamlParsing)
		}

		// Create the route ...
		_, errRoute := routeV1Client.Routes(application.Namespace).Create(&route)
		if errRoute != nil {
			log.Fatal("error creating route", errRoute.Error())
		}
	}
}

func getClient(config *restclient.Config) *routeclientsetv1.RouteV1Client {
	routeV1Client, errrouteclientsetv1 := routeclientsetv1.NewForConfig(config)
	if errrouteclientsetv1 != nil {
		log.Fatal("error creating route Clientset", errrouteclientsetv1.Error())
	}
	return routeV1Client
}

func DeleteRoute(config *restclient.Config, application types.Application) {
	if oc.Exists("route", application.Name) {
		// Create the route ...
		errRoute := getClient(config).Routes(application.Namespace).Delete(application.Name, deleteOptions)
		if errRoute != nil {
			log.Fatalf("Unable to delete Route: %s", errRoute.Error())
		}
	}
}
