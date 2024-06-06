package flags

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
		ip, port := utils.ValidateIPAndPort(serverIP, serverPort)
		fmt.Printf("Starting server at %s:%d with files: %v\n", ip, port, serverFiles)
		srv := server.NewServer(ip, port)
		srv.ServerRun(serverFiles)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVarP(&serverIP, "ip", "i", "", "IP Address (default \"Ip currently in use\")")
	serverCmd.Flags().IntVarP(&serverPort, "port", "p", 8877, "Port Number")
	serverCmd.Flags().StringArrayVarP(&serverFiles, "files", "f", []string{}, "Files to serve")
}
