package log

import (
	"fmt"
	"os"
	"sync"
)

var PrintMutex sync.Mutex

func PrintToOut(format string, args ...any) {
	PrintMutex.Lock()
	defer PrintMutex.Unlock()
	fmt.Fprintf(os.Stdout, format, args...)
}

func PrintToErr(format string, args ...any) {
	PrintMutex.Lock()
	defer PrintMutex.Unlock()
	fmt.Fprintf(os.Stderr, format, args...)
}
