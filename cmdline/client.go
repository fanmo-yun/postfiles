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
		ip, port, validateErr := utils.ValidateIPAndPort(clientIP, clientPort)
		if validateErr != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error validating IP and port: %s\n", validateErr)
			return
		}
		if clientSavePath == "System Download Path" {
			sp, getErr := utils.GetDownloadPath()
			if getErr != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Failed to get download path: %s\n", getErr)
				return
			}
			clientSavePath = sp
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Starting client at %s:%d, saving to %s\n", ip, port, clientSavePath)
		client := client.NewClient(ip, port, clientSavePath)
		if validateErr := client.ValidateSavePath(); validateErr != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Failed to validate save path: %s\n", validateErr)
			return
		}
		if connectErr := client.Start(); connectErr != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Failed to connect to server: %s\n", connectErr)
			return
		}
	},
}
