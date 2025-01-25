package utils

import "fmt"

const (
	unitKB = 1024
	unitMB = 1024 * unitKB
	unitGB = 1024 * unitMB
)

func ToReadableSize(size int64) string {
	switch {
	case size >= unitGB:
		return fmt.Sprintf("%.2f GB", float64(size)/unitGB)
	case size >= unitMB:
		return fmt.Sprintf("%.2f MB", float64(size)/unitMB)
	case size >= unitKB:
		return fmt.Sprintf("%.2f KB", float64(size)/unitKB)
	default:
		return fmt.Sprintf("%d B", size)
	}
}
