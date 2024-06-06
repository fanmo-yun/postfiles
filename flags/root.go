package flags

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "postfiles",
	Short: "Postfiles is a tool for file transfer",
	Long:  `Postfiles is a CLI tool to serve and receive files over a network.`,
}

func Execute() {
	if executeErr := rootCmd.Execute(); executeErr != nil {
		fmt.Fprintf(os.Stderr, "%v", executeErr)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize()
}
