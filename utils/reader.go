package utils

import (
	"bufio"
	"os"
	"strings"
)

func Readin() (string, error) {
	reader := bufio.NewReaderSize(os.Stdin, 16)
	confirm, readErr := reader.ReadString('\n')
	if readErr != nil {
		return "", readErr
	}
	return strings.TrimRight(confirm, "\r\n"), nil
}
