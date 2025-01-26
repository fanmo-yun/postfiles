package utils

import (
	"os"
	"unicode/utf8"

	"github.com/schollz/progressbar/v3"
	"golang.org/x/term"
)

func CreateProcessBar(filesize int64, filename string) *progressbar.ProgressBar {
	// w := GetBarWidth()
	processbar := progressbar.NewOptions64(
		filesize,
		progressbar.OptionSetDescription(filename),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionFullWidth(),
	)
	return processbar
}

func GetBarWidth() int {
	width, _, _ := term.GetSize(int(os.Stdout.Fd()))
	barLength := width * 70 / 100
	if barLength < 20 {
		barLength = 20
	} else if barLength > width-10 {
		barLength = width - 10
	}
	return barLength
}

func TruncateString(s string, maxLength int) string {
	if utf8.RuneCountInString(s) < maxLength {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxLength-3]) + "..."
}
