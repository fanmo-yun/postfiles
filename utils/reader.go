package utils

import (
	"bufio"
	"fmt"
	"os"
	"postfiles/exitcodes"
	"strings"
)

func Readin() string {
	reader := bufio.NewReader(os.Stdin)
	confirm, readErr := reader.ReadString('\n')
	if readErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to read input: %s\n", readErr)
		os.Exit(exitcodes.ErrReadInput)
	}
	return strings.TrimRight(confirm, "\r\n")
}
