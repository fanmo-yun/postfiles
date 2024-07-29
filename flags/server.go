package flags

import (
	"fmt"
	"postfiles/server"
	"postfiles/utils"

	"github.com/spf13/cobra"
)

var (
	serverIP    string
	serverPort  int
	serverFiles []string
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run as server",
	Run: func(cmd *cobra.Command, args []string) {
		if len(serverFiles) == 0 {
			fmt.Fprintf(cmd.OutOrStderr(), "No files provided to serve.\n")
			cmd.Usage()
			return
		}

		ip, port := utils.ValidateServerIPAndPort(serverIP, serverPort)
		fmt.Fprintf(cmd.OutOrStderr(), "Starting server at %s:%d\n", ip, port)
		srv := server.NewServer(ip, port)
		srv.ServerRun(serverFiles)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVarP(&serverIP, "ip", "i", "", "IP Address (default \"Ip currently in use\")")
	serverCmd.Flags().IntVarP(&serverPort, "port", "p", 8877, "Port Number")
	serverCmd.Flags().StringSliceVarP(&serverFiles, "files", "f", []string{}, "Files to serve")
}
