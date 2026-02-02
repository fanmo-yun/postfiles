package cmdline

import (
	"log/slog"

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
	rootCmd.AddCommand(clientCmd)
	rootCmd.AddCommand(versionCmd)

	serverCmd.Flags().StringVarP(&serverIP, "ip", "i", "", "IP Address (default: auto)")
	serverCmd.Flags().IntVarP(&serverPort, "port", "p", 8877, "Port Number")
	serverCmd.Flags().StringSliceVarP(&serverFiles, "file", "f", nil, "Files to serve")

	clientCmd.Flags().StringVarP(&clientIP, "ip", "i", "", "IP Address (default: auto)")
	clientCmd.Flags().IntVarP(&clientPort, "port", "p", 8877, "Port Number")
	clientCmd.Flags().StringVarP(&clientSavePath, "save", "s", "", "Save Path")

	if err := rootCmd.Execute(); err != nil {
		slog.Error("cli execute failed", "err", err)
		return
	}
}
