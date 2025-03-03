package cmdline

import (
	"postfiles/log"
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
			log.PrintToErr("Server: No files provided to serve\n")
			cmd.Usage()
			return
		}

		ip, port, validateErr := utils.ValidateIPAndPort(serverIP, serverPort)
		if validateErr != nil {
			log.PrintToErr("Server: Error validating IP and port\n")
			return
		}
		server := server.NewServer(ip, port, serverFiles)
		if err := server.Start(); err != nil {
			log.PrintToErr("Server Fatal: %s\n", err)
			return
		}
	},
}
