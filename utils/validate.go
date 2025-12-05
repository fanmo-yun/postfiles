package utils

import (
	"errors"
	"net"
	"os"

	"golang.org/x/term"
)

func ipok(ip string) error {
	if parseErr := net.ParseIP(ip); parseErr == nil {
		return errors.New("ip incorrect")
	}
	return nil
}

func portok(port int) error {
	if port < 1024 || port > 65535 {
		return errors.New("port incorrect")
	}
	return nil
}

func Check(ip string, port int) (string, int, error) {
	if ip == "" {
		autoIP, err := GenIP()
		if err != nil {
			return "", 0, err
		}
		ip = autoIP
	}

	if err := ipok(ip); err != nil {
		return "", 0, err
	}

	if err := portok(port); err != nil {
		return "", 0, err
	}

	return ip, port, nil
}

func IsTerm() error {
	stdout := term.IsTerminal(int(os.Stdout.Fd()))
	stderr := term.IsTerminal(int(os.Stderr.Fd()))
	stdin := term.IsTerminal(int(os.Stdin.Fd()))

	if stdout && stderr && stdin {
		return nil
	}
	return errors.New("not a terminal")
}
