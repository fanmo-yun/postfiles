package flags

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
		ip, port := utils.ValidateClientIPAndPort(clientIP, clientPort)
		if clientSavePath == "System Download Path" {
			clientSavePath = utils.GetDownloadPath()
		}
		fmt.Printf("Starting client at %s:%d, saving to %s\n", ip, port, clientSavePath)
		clt := client.NewClient(ip, port)
		clt.ClientRun(clientSavePath)
	},
}

func init() {
	rootCmd.AddCommand(clientCmd)
	clientCmd.Flags().StringVarP(&clientIP, "ip", "i", "", "IP Address (default \"Ip currently in use\")")
	clientCmd.Flags().IntVarP(&clientPort, "port", "p", 8877, "Port Number")
	clientCmd.Flags().StringVarP(&clientSavePath, "save", "s", "System Download Path", "Save Path")
}
