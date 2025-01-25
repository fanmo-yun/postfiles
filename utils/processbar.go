package utils

import (
	"fmt"
	"unicode/utf8"

	"github.com/schollz/progressbar/v3"
)

func CreateBar(filesize int64, filename string) *progressbar.ProgressBar {
	w := GetBarWidth()
	bar := progressbar.NewOptions64(
		filesize,
		progressbar.OptionSetDescription(TruncateString(filename, w-65)),
		progressbar.OptionSetWidth(w-35),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Println()
		}),
		progressbar.OptionSetRenderBlankState(true),
	)
	return bar
}

func TruncateString(s string, maxLength int) string {
	if utf8.RuneCountInString(s) < maxLength {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxLength-3]) + "..."
}
