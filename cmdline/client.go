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
			sp, getErr := utils.GetDownloadDir()
			if getErr != nil {
				slog.Error("failed to get download path", "err", getErr)
				return
			}
			clientSavePath = sp
		}

		saveDir := filepath.Clean(clientSavePath)
		root, rootErr := os.OpenRoot(saveDir)
		if rootErr != nil {
			slog.Error("failed to open root", "err", rootErr)
			return
		}

		address := net.JoinHostPort(ip, strconv.Itoa(port))
		client := client.NewClient(address, root)
		if validateErr := client.ValidateWritable(); validateErr != nil {
			slog.Error("failed to validate save path", "err", validateErr)
			return
		}
		if connectErr := client.Start(); connectErr != nil {
			slog.Error("client fatal", "err", connectErr)
			return
		}
	},
}
