package flags

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStderr(), "v1.2.0\n")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
