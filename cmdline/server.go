package cmdline

import (
	"postfiles/server"
	"postfiles/utils"

	"github.com/rs/zerolog/log"
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
			log.Error().Msg("No files provided to serve")
			cmd.Usage()
			return
		}

		ip, port, validateErr := utils.ValidateIPAndPort(serverIP, serverPort)
		if validateErr != nil {
			log.Error().Err(validateErr).Msg("Error validating IP and port")
		}
		server := server.NewServer(ip, port, serverFiles)
		if err := server.Start(); err != nil {
			log.Error().Err(err).Msg("Error starting server")
			return
		}
	},
}
