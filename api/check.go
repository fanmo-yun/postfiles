package api

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
