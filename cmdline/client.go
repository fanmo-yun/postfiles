package cmdline

import (
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"postfiles/client"
	"postfiles/utils"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	clientIP       string
	clientPort     int
	clientSavePath string
)

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Run as client",
	Run: func(cmd *cobra.Command, args []string) {
		ip, port, err := utils.Check(clientIP, clientPort)
		if err != nil {
			slog.Error("validating IP and port failed", "err", err)
			return
		}
		if clientSavePath == "" {
			sp, err := utils.GetDownloadDir()
			if err != nil {
				slog.Error("failed to get download path", "err", err)
				return
			}
			clientSavePath = sp
		}

		saveDir := filepath.Clean(clientSavePath)
		root, err := os.OpenRoot(saveDir)
		if err != nil {
			slog.Error("failed to open root", "err", err)
			return
		}

		address := net.JoinHostPort(ip, strconv.Itoa(port))
		client := client.NewClient(address, root)
		if err := client.ValidateWritable(); err != nil {
			slog.Error("failed to validate save path", "err", err)
			return
		}
		if err := client.Start(); err != nil {
			slog.Error("client fatal", "err", err)
			return
		}
	},
}
