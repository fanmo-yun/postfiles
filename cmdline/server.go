package cmdline

import (
	"log/slog"
	"net"
	"postfiles/server"
	"postfiles/utils"
	"strconv"

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
			slog.Error("server: no files provided to serve")
			if err := cmd.Usage(); err != nil {
				slog.Error("server: usage print failed", "err", err)
				return
			}
			return
		}

		ip, port, err := utils.Check(serverIP, serverPort)
		if err != nil {
			slog.Error("validating IP and port failed", "err", err)
			return
		}

		address := net.JoinHostPort(ip, strconv.Itoa(port))
		server := server.NewServer(address, serverFiles)
		if err := server.Start(); err != nil {
			slog.Error("server fatal", "err", err)
			return
		}
	},
}
