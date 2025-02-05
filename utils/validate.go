package utils

import (
	"errors"
	"net"
	"os"

	"golang.org/x/term"
)

func ValidIP(ip string) error {
	if parsErr := net.ParseIP(ip); parsErr == nil {
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

func ValidateIPAndPort(ip string, port int) (string, int, error) {
	if len(ip) == 0 {
		temp_ip, genErr := GenIP()
		if genErr != nil {
			return "", 0, genErr
		}
		ip = temp_ip
	}
	if ipErr := ValidIP(ip); ipErr != nil {
		return "", 0, ipErr
	}

	if portErr := ValidPort(port); portErr != nil {
		return "", 0, portErr
	}

	return ip, port, nil
}

func IsTerminal() error {
	if !(term.IsTerminal(int(os.Stdout.Fd())) &&
		term.IsTerminal(int(os.Stderr.Fd())) &&
		term.IsTerminal(int(os.Stdin.Fd()))) {
		return errors.New("not a terminal")
	}
	return nil
}
