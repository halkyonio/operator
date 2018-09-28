package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"fmt"
	"github.com/snowdrop/spring-boot-cloud-devex/pkg/common/config"
	"github.com/snowdrop/spring-boot-cloud-devex/pkg/common/oc"
	"strings"
)

func newCommand(action string) *cobra.Command {
	return newCommandWith(action, execAction)
}

func newCommandWith(action string, toExec func(podName string, action string)) *cobra.Command {
	capitalizedAction := strings.Title(action)

	return &cobra.Command{
		Use:     action,
		Short:   capitalizedAction + " your SpringBoot application.",
		Long:    capitalizedAction + ` your SpringBoot application.`,
		Example: `  sb exec ` + action,
		Args:    cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {

			log.Infof("Exec %s command called", action)

			_, pod := SetupAndWaitForPod()
			podName := pod.Name

			log.Infof("%s the Spring Boot application ...", capitalizedAction)
			toExec(podName, action)
			oc.ExecCommand(oc.Command{Args: []string{"logs", podName, "-f"}})
		},
	}
}

func execAction(podName string, action string) {
	cmdArgs := []string{"rsh", podName, config.SupervisordBin, config.SupervisordCtl, action, config.RunCmdName}
	log.Debug("Command :", cmdArgs)
	oc.ExecCommand(oc.Command{Args: cmdArgs})
}

func init() {
	var ports string

	execStartCmd := newCommand("start")
	execStopCmd := newCommand("stop")
	execRestartCmd := newCommandWith("restart", func(podName string, action string) {
		oc.ExecCommand(oc.Command{Args: []string{"rsh", podName, config.SupervisordBin, config.SupervisordCtl, "stop", config.RunCmdName}})
		oc.ExecCommand(oc.Command{Args: []string{"rsh", podName, config.SupervisordBin, config.SupervisordCtl, "start", config.RunCmdName}})
	})
	execDebugCmd := newCommandWith("debug", func(podName string, action string) {
		oc.ExecCommand(oc.Command{Args: []string{"rsh", podName, config.SupervisordBin, config.SupervisordCtl, "stop", config.RunCmdName}})
		oc.ExecCommand(oc.Command{Args: []string{"rsh", podName, config.SupervisordBin, config.SupervisordCtl, "start", config.RunCmdName}})

		// Forward local to Remote port
		log.Info("Remote Debug the Spring Boot Application ...")
		oc.ExecCommand(oc.Command{Args: []string{"port-forward", podName, ports}})
	})

	execDebugCmd.Flags().StringVarP(&ports, "ports", "p", "5005:5005", "Local and remote ports to be used to forward traffic between the dev pod and your machine.")
	execDebugCmd.Example = "  sb exec debug -p5005:8080"

	exeCmd := &cobra.Command{
		Use:   "exec [options]",
		Short: "Stop, start, debug or restart your SpringBoot application.",
		Long:  `Stop, start, debug  or restart your SpringBoot application.`,
		Example: fmt.Sprintf("%s\n%s\n%s\n%s",
			execStartCmd.Example,
			execStopCmd.Example,
			execRestartCmd.Example,
			execDebugCmd.Example),
	}
	exeCmd.AddCommand(execStartCmd, execStopCmd, execRestartCmd, execDebugCmd)

	exeCmd.Annotations = map[string]string{"command": "exec"}
	rootCmd.AddCommand(exeCmd)
}
