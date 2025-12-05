package cmdline

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	BuildVersion = "dev"
	BuildCommit  = "none"
	BuildTime    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("postfiles")
		fmt.Printf("  Version: %s\n", BuildVersion)
		fmt.Printf("  Commit:  %s\n", BuildCommit)
		fmt.Printf("  Built:   %s\n", BuildTime)
	},
}
