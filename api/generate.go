package api

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func GenIP() string {
	conn, connErr := net.Dial("udp", "114.114.114.114:80")
	if connErr != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to UDP server: %v\n", connErr)
		os.Exit(1)
	}
	defer conn.Close()
	return strings.Split(conn.LocalAddr().String(), ":")[0]
}

func GetDownloadPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		os.Exit(1)
	}
	downloadDir := filepath.Join(homeDir, "Downloads")
	if _, dirErr := os.Stat(downloadDir); os.IsNotExist(dirErr) {
		fmt.Fprintf(os.Stderr, "Download directory does not exist: %s\n", downloadDir)
		os.Exit(1)
	} else if dirErr != nil {
		fmt.Fprintf(os.Stderr, "Error stating download directory: %v\n", dirErr)
		os.Exit(1)
	}
	return downloadDir
}
