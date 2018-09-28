package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	// VERSION is set during build
	VERSION string

	// GITCOMMIT is hash of the commit that wil be displayed when running ./odo version
	// this will be overwritten when running  build like this: go build -ldflags="-X github.com/redhat-developer/odo/cmd.GITCOMMIT=$(GITCOMMIT)"
	// HEAD is default indicating that this was not set during build
	GITCOMMIT = "HEAD"
)

func init() {
	versionCmd := &cobra.Command{
		Use:     "version",
		Short:   "Show sb client version",
		Long:    `Show sb client version.`,
		Example: ` sb version`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(rootCmd.Use + " v" + VERSION + " (" + GITCOMMIT + ")")
		},
	}

	rootCmd.AddCommand(versionCmd)
}
