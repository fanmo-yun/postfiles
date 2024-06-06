package utils

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
)

func IsvalidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

func IsvalidPort(port int) bool {
	return port >= 1024 && port <= 65535
}

func FileStat(file string) (string, int64) {
	info, statErr := os.Stat(file)

	if os.IsNotExist(statErr) {
		fmt.Fprintf(os.Stderr, "file: %s does not exist\n", file)
		os.Exit(1)
	} else if info.IsDir() {
		fmt.Fprintf(os.Stderr, "file: %s is a directory\n", file)
		os.Exit(1)
	} else if statErr != nil {
		fmt.Fprintf(os.Stderr, "error stating file: %v\n", statErr)
		os.Exit(1)
	}

	return filepath.Base(info.Name()), info.Size()
}

func ValidateIPAndPort(ip string, port int) (string, int) {
	temp_ip := ip
	temp_port := port

	if len(temp_ip) == 0 {
		temp_ip = GenIP()
	} else if !IsvalidIP(temp_ip) {
		fmt.Fprintf(os.Stderr, "Invalid IP: %s\n", temp_ip)
		os.Exit(1)
	}

	if !IsvalidPort(temp_port) {
		fmt.Fprintf(os.Stderr, "Invalid Port: %d\n", temp_port)
		os.Exit(1)
	}

	return temp_ip, temp_port
}
