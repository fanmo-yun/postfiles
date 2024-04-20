package api

import (
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func GenIP() string {
	conn, connErr := net.Dial("udp", "114.114.114.114:80")
	if connErr != nil {
		log.Fatal(connErr)
	}
	defer conn.Close()
	return strings.Split(conn.LocalAddr().String(), ":")[0]
}

func GetDownloadPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	downloadDir := filepath.Join(homeDir, "Downloads")
	if _, dirErr := os.Stat(downloadDir); os.IsNotExist(dirErr) {
		log.Fatal("download directory does not exist")
	}
	return downloadDir
}
