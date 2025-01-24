package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func Readin() string {
	reader := bufio.NewReaderSize(os.Stdin, 16)
	confirm, readErr := reader.ReadString('\n')
	if readErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to read input: %s\n", readErr)
		os.Exit(ErrReadInput)
	}
	return strings.TrimRight(confirm, "\r\n")
}
