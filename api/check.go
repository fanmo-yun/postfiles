package api

import (
	"net"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
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
		logrus.Fatalf("file: %s does not exist", file)
	} else if info.IsDir() {
		logrus.Fatalf("error stating file: %v", statErr)
	}
	return filepath.Base(info.Name()), info.Size()
}
