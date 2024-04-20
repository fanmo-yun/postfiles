package api

import (
	"log"
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
		log.Fatalf("file: %s does not exist\n", file)
	} else if info.IsDir() {
		log.Fatalf("%s is dir\n", file)
	}
	return filepath.Base(info.Name()), info.Size()
}
