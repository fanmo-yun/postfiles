package utils

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
)

func GenIP() (string, error) {
	conn, err := net.Dial("udp", "114.114.114.114:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	addr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return "", errors.New("unexpected local address type")
	}
	return addr.IP.String(), nil
}

func GetDownloadDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(home, "Downloads")
	info, statErr := os.Stat(dir)

	if statErr != nil {
		if os.IsNotExist(statErr) {
			return "", fmt.Errorf("%s does not exist", dir)
		}
		return "", statErr
	}

	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", dir)
	}

	return dir, nil
}
