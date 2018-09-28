package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/snowdrop/spring-boot-cloud-devex/pkg/common/config"
	"github.com/snowdrop/spring-boot-cloud-devex/pkg/common/oc"
)

func init() {
	compileCmd := &cobra.Command{
		Use:     "compile",
		Short:   "Compile local project within the development pod",
		Long:    `Compile local project within the development pod.`,
		Example: ` sb compile`,
		Args:    cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {

			log.Info("Compile command called")

			_, pod := SetupAndWaitForPod()
			podName := pod.Name

			log.Info("Compile ...")
			oc.ExecCommand(oc.Command{Args: []string{"rsh", podName, config.SupervisordBin, config.SupervisordCtl, "start", config.CompileCmdName}})
			oc.ExecCommand(oc.Command{Args: []string{"logs", podName, "-f"}})
		},
	}

	compileCmd.Annotations = map[string]string{"command": "compile"}
	rootCmd.AddCommand(compileCmd)
}
