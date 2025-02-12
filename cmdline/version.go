package cmdline

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	cliName    = "postfiles"
	cliVersion = "v1.2.2"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "%s - %s\n", cliName, cliVersion)
	},
}
