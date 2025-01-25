package utils

import (
	"fmt"

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
