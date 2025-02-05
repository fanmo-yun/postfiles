package utils

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func GenIP() (string, error) {
	conn, connErr := net.Dial("udp", "114.114.114.114:80")
	if connErr != nil {
		return "", connErr
	}
	defer conn.Close()
	return strings.Split(conn.LocalAddr().String(), ":")[0], nil
}

func GetDownloadPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	downloadDir := filepath.Join(homeDir, "Downloads")
	if dirStat, dirErr := os.Stat(downloadDir); os.IsNotExist(dirErr) || !dirStat.IsDir() {
		return "", fmt.Errorf("[%s] path does not exist or is not a folder", downloadDir)
	} else if dirErr != nil {
		return "", dirErr
	}
	return downloadDir, nil
}
