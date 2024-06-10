package client

import (
	"fmt"
	"postfiles/utils"

	"github.com/schollz/progressbar/v3"
)

func init_bar(filesize int64, filename string) *progressbar.ProgressBar {
	w := utils.GetBarWidth()
	bar := progressbar.NewOptions64(
		filesize,
		progressbar.OptionSetDescription(utils.TruncateString(filename, w-65)),
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
