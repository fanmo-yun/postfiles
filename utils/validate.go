package utils

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"golang.org/x/term"
)

func ValidIP(ip string) error {
	if parsErr := net.ParseIP(ip); parsErr != nil {
		return errors.New("ip incorrect")
	}
	return nil
}

func ValidPort(port int) error {
	if !(port >= 1024 && port <= 65535) {
		return errors.New("port incorrect")
	}
	return nil
}

func FileStat(file string) (string, int64) {
	info, statErr := os.Stat(file)

	if os.IsNotExist(statErr) {
		fmt.Fprintf(os.Stderr, "file: %s does not exist\n", file)
		os.Exit(ErrFileStat)
	} else if info.IsDir() {
		fmt.Fprintf(os.Stderr, "file: %s is a directory\n", file)
		os.Exit(ErrFileStat)
	} else if statErr != nil {
		fmt.Fprintf(os.Stderr, "error stating file: %s\n", statErr)
		os.Exit(ErrFileStat)
	}

	return filepath.Base(info.Name()), info.Size()
}

func ValidateIPAndPort(ip string, port int) (string, int) {
	temp_ip := ip
	temp_port := port

	if len(temp_ip) == 0 {
		temp_ip = GenIP()
	}
	if ipErr := ValidIP(temp_ip); ipErr != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", ipErr, temp_ip)
		os.Exit(ErrIPAndPort)
	}

	if portErr := ValidPort(temp_port); portErr != nil {
		fmt.Fprintf(os.Stderr, "%s: %d\n", portErr, temp_port)
		os.Exit(ErrIPAndPort)
	}

	return temp_ip, temp_port
}

func IsTerminal() {
	if !(term.IsTerminal(int(os.Stdout.Fd())) &&
		term.IsTerminal(int(os.Stderr.Fd())) &&
		term.IsTerminal(int(os.Stdin.Fd()))) {
		fmt.Fprintf(os.Stderr, "Not in a terminal\n")
		os.Exit(ErrNotTerminal)
	}
}
