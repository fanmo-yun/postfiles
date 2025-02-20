package cmdline

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
	cobra.EnableCommandSorting = false
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVarP(&serverIP, "ip", "i", "", "IP Address (default \"Ip currently in use\")")
	serverCmd.Flags().IntVarP(&serverPort, "port", "p", 8877, "Port Number")
	serverCmd.Flags().StringSliceVarP(&serverFiles, "file", "f", make([]string, 0, 10), "Files to serve")

	rootCmd.AddCommand(clientCmd)
	clientCmd.Flags().StringVarP(&clientIP, "ip", "i", "", "IP Address (default \"Ip currently in use\")")
	clientCmd.Flags().IntVarP(&clientPort, "port", "p", 8877, "Port Number")
	clientCmd.Flags().StringVarP(&clientSavePath, "save", "s", "System Download Path", "Save Path")

	rootCmd.AddCommand(versionCmd)

	if executeErr := rootCmd.Execute(); executeErr != nil {
		fmt.Fprintf(os.Stderr, "%s\n", executeErr)
		return
	}
}
