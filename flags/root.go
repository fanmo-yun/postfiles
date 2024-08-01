package flags

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "postfiles",
	Short: "PostFiles is a tool for file transfer.",
	Long:  "PostFiles is a CLI tool for transferring files over TCP.",
}

func Execute() {
	if executeErr := rootCmd.Execute(); executeErr != nil {
		fmt.Fprintf(os.Stderr, "%s", executeErr)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize()
}
