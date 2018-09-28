package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/openshift/api/apps/v1"
	"github.com/snowdrop/spring-boot-cloud-devex/pkg/buildpack"
	"strings"
)

func init() {
	initCmd := &cobra.Command{
		Use:     "init [flags]",
		Short:   "Create a development's pod for the component",
		Long:    `Create a development's pod for the component.`,
		Example: ` sb init -n bootapp`,
		Args:    cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {

			log.Info("Init command called")
			setup := Setup()

			cmds := ""
			if len(args) > 0 && strings.HasPrefix(args[0], "CMDS") {
				cmds = strings.TrimPrefix(args[0], "CMDS=")
			}

			log.Info("Commands to be passed to the supervisord : ", cmds)

			// Create ImageStreams
			log.Info("Create ImageStreams for Supervisord and Java S2I Image of SpringBoot")
			buildpack.CreateDefaultImageStreams(setup.RestConfig, setup.Application)

			// Create PVC
			log.Info("Create PVC to store m2 repo")
			buildpack.CreatePVC(setup.Clientset, setup.Application, "1Gi")

			var dc *v1.DeploymentConfig
			log.Info("Create or retrieve DeploymentConfig using Supervisord and Java S2I Image of SpringBoot")
			dc = buildpack.CreateOrRetrieveDeploymentConfig(setup.RestConfig, setup.Application, cmds)

			log.Info("Create Service using Template")
			buildpack.CreateServiceTemplate(setup.Clientset, dc, setup.Application)

			log.Info("Create Route using Template")
			buildpack.CreateRouteTemplate(setup.RestConfig, setup.Application)
		},
	}

	// Add a defined annotation in order to appear in the help menu
	initCmd.Annotations = map[string]string{"command": "init"}

	rootCmd.AddCommand(initCmd)
}
