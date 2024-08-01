package utils

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/term"
)

func GenIP() string {
	conn, connErr := net.Dial("udp", "114.114.114.114:80")
	if connErr != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to UDP server: %s\n", connErr)
		os.Exit(1)
	}
	defer conn.Close()
	return strings.Split(conn.LocalAddr().String(), ":")[0]
}

func GetDownloadPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %s\n", err)
		os.Exit(1)
	}
	downloadDir := filepath.Join(homeDir, "Downloads")
	if _, dirErr := os.Stat(downloadDir); os.IsNotExist(dirErr) {
		fmt.Fprintf(os.Stderr, "Download directory does not exist: %s\n", downloadDir)
		os.Exit(1)
	} else if dirErr != nil {
		fmt.Fprintf(os.Stderr, "Error stating download directory: %s\n", dirErr)
		os.Exit(1)
	}
	return downloadDir
}

func GetBarWidth() int {
	width, _, _ := term.GetSize(int(os.Stdout.Fd()))
	barLength := width * 70 / 100
	if barLength < 20 {
		barLength = 20
	} else if barLength > width-10 {
		barLength = width - 10
	}
	return barLength
}
