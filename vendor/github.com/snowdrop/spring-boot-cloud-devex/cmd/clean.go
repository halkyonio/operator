package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/snowdrop/spring-boot-cloud-devex/pkg/buildpack"
)

func init() {
	cleanCmd := &cobra.Command{
		Use:     "clean [flags]",
		Short:   "Remove development pod for the component",
		Long:    `Remove development pod for the component.`,
		Example: ` sb clean`,
		Args:    cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {

			log.Info("Clean command called")

			setup := Setup()

			buildpack.DeleteDefaultImageStreams(setup.RestConfig, setup.Application)

			buildpack.DeletePVC(setup.Clientset, setup.Application)

			buildpack.DeleteDeploymentConfig(setup.RestConfig, setup.Application)

			buildpack.DeleteService(setup.Clientset, setup.Application)

			buildpack.DeleteRoute(setup.RestConfig, setup.Application)

			log.Info("Deleted resources")
		},
	}

	// Add a defined annotation in order to appear in the help menu
	cleanCmd.Annotations = map[string]string{"command": "clean"}

	rootCmd.AddCommand(cleanCmd)
}
