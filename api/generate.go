package api

import (
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

func GenIP() string {
	conn, connErr := net.Dial("udp", "114.114.114.114:80")
	if connErr != nil {
		logrus.Fatal(connErr)
	}
	defer conn.Close()
	return strings.Split(conn.LocalAddr().String(), ":")[0]
}

func GetDownloadPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logrus.Fatal(err)
	}
	downloadDir := filepath.Join(homeDir, "Downloads")
	if _, dirErr := os.Stat(downloadDir); os.IsNotExist(dirErr) {
		logrus.Fatal("download directory does not exist")
	}
	return downloadDir
}
