package cmdline

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

		ip, port, validateErr := utils.ValidateIPAndPort(serverIP, serverPort)
		if validateErr != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error validating IP and port: %s\n", validateErr)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Starting server at %s:%d\n", ip, port)
		server := server.NewServer(ip, port, serverFiles)
		if err := server.Start(); err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error starting server: %s\n", err)
			return
		}
	},
}
