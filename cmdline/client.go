package cmdline

import (
	"postfiles/client"
	"postfiles/log"
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
		ip, port, validateErr := utils.ValidateIPAndPort(clientIP, clientPort)
		if validateErr != nil {
			log.PrintToErr("Error validating IP and port: %s\n", validateErr)
			return
		}
		if clientSavePath == "System Download Path" {
			sp, getErr := utils.GetDownloadPath()
			if getErr != nil {
				log.PrintToErr("Failed to get download path: %s\n", getErr)
				return
			}
			clientSavePath = sp
		}
		client := client.NewClient(ip, port, clientSavePath)
		if validateErr := client.ValidateSavePath(); validateErr != nil {
			log.PrintToErr("Failed to validate save path: %s\n", validateErr)
			return
		}
		if connectErr := client.Start(); connectErr != nil {
			log.PrintToErr("Client Fatal: %s\n", connectErr)
			return
		}
	},
}
