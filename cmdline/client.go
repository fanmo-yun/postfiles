package cmdline

import (
	"fmt"
	"postfiles/client"
	"postfiles/utils"

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
		ip, port := utils.ValidateIPAndPort(clientIP, clientPort)
		if clientSavePath == "System Download Path" {
			clientSavePath = utils.GetDownloadPath()
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Starting client at %s:%d, saving to %s\n", ip, port, clientSavePath)
		client := client.NewClient(ip, port, clientSavePath)
		client.ValidateSavePath()
		client.ClientRun()
	},
}
