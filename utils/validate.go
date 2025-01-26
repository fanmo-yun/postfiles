package utils

import (
	"errors"
	"fmt"
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

func ValidateIPAndPort(ip string, port int) (string, int) {
	temp_ip := ip

	if len(temp_ip) == 0 {
		temp_ip = GenIP()
	}
	if ipErr := ValidIP(temp_ip); ipErr != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", ipErr, temp_ip)
		os.Exit(ErrIPAndPort)
	}

	if portErr := ValidPort(port); portErr != nil {
		fmt.Fprintf(os.Stderr, "%s: %d\n", portErr, port)
		os.Exit(ErrIPAndPort)
	}

	return temp_ip, port
}

func IsTerminal() {
	if !(term.IsTerminal(int(os.Stdout.Fd())) &&
		term.IsTerminal(int(os.Stderr.Fd())) &&
		term.IsTerminal(int(os.Stdin.Fd()))) {
		fmt.Fprintf(os.Stderr, "Not in a terminal\n")
		os.Exit(ErrNotTerminal)
	}
}
