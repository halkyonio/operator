package buildpack

import (
	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"

	appsv1 "github.com/openshift/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/snowdrop/spring-boot-cloud-devex/pkg/buildpack/types"
	"github.com/snowdrop/spring-boot-cloud-devex/pkg/common/oc"
	"strings"
)

func CreateServiceTemplate(clientset *kubernetes.Clientset, dc *appsv1.DeploymentConfig, application types.Application) {
	if oc.Exists("svc", application.Name) {
		log.Infof("'%s' Service already exists, skipping", application.Name)
	} else {
		// Parse Service Template
		tName := strings.Join([]string{builderPath, "service"}, "/")
		var b = ParseTemplate(tName, application)

		// Create Service struct using the generated Service string
		svc := corev1.Service{}
		errYamlParsing := yaml.Unmarshal(b.Bytes(), &svc)
		if errYamlParsing != nil {
			panic(errYamlParsing)
		}
		_, errService := clientset.CoreV1().Services(application.Namespace).Create(&svc)
		if errService != nil {
			log.Fatalf("Unable to create Service: %s", errService.Error())
		}
	}
}

func DeleteService(clientset *kubernetes.Clientset, application types.Application) {
	if oc.Exists("svc", application.Name) {
		errService := clientset.CoreV1().Services(application.Namespace).Delete(application.Name, deleteOptions)
		if errService != nil {
			log.Fatalf("Unable to delete Service: %s", errService.Error())
		}
	}
}
