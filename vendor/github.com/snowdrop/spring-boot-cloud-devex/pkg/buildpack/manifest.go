package buildpack

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"

	"encoding/json"
	"github.com/ghodss/yaml"
	"github.com/snowdrop/spring-boot-cloud-devex/pkg/buildpack/types"
	"os"
)

func ParseManifest(manifestPath string) types.Application {
	log.Debugf("Parsing Application Config at %s", manifestPath)

	// Create an Application with default values
	appConfig := types.NewApplication()

	// if we have a manifest file, use it to replace default values
	if _, err := os.Stat(manifestPath); err == nil {
		source, err := ioutil.ReadFile(manifestPath)
		if err != nil {
			panic(err)
		}

		err = yaml.Unmarshal(source, &appConfig)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Infof("No MANIFEST file detected, using default values")
	}

	log.Infof("Application configured")

	if log.GetLevel() == log.DebugLevel {
		log.Debug("Application's config")
		log.Debug("--------------------")
		appFormatted, _ := json.Marshal(appConfig)
		log.Debug(string(appFormatted))
	}

	return appConfig
}
