package flags

import (
	"fmt"
	"os"
	"postfiles/exitcodes"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "postfiles",
	Short: "PostFiles is a tool for file transfer.",
	Long:  "PostFiles is a CLI tool for transferring files over TCP.",
}

func Execute() {
	cobra.EnableCommandSorting = false

	if executeErr := rootCmd.Execute(); executeErr != nil {
		fmt.Fprintf(os.Stderr, "%s\n", executeErr)
		os.Exit(exitcodes.ErrFlag)
	}
}

func init() {
	cobra.OnInitialize()
}
