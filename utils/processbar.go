package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/schollz/progressbar/v3"
	"golang.org/x/term"
	"golang.org/x/text/width"
)

func CreateProcessBar(filesize int64, filename string) *progressbar.ProgressBar {
	barWidth := GetBarWidth()
	textWidth := barWidth * 30 / 100
	afterText := PadOrTruncateString(filename, textWidth)

	processbar := progressbar.NewOptions64(
		filesize,
		progressbar.OptionSetDescription(afterText),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprintf(os.Stdout, "\n")
		}),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionFullWidth(),
	)
	return processbar
}

func GetBarWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 80
	}
	return width
}

func GetTextWidth(s string) int {
	w := 0
	for _, r := range s {
		switch width.LookupRune(r).Kind() {
		case width.EastAsianFullwidth, width.EastAsianWide:
			w += 2
		case width.EastAsianHalfwidth, width.EastAsianNarrow,
			width.Neutral, width.EastAsianAmbiguous:
			w += 1
		}
	}
	return w
}

func PadOrTruncateString(s string, targetLength int) string {
	currentWidth := GetTextWidth(s)
	builder := new(strings.Builder)

	if currentWidth > targetLength {
		w := 0
		for _, r := range s {
			runeWidth := 1
			switch width.LookupRune(r).Kind() {
			case width.EastAsianFullwidth, width.EastAsianWide:
				runeWidth = 2
			}

			if w+runeWidth > targetLength-3 {
				break
			}
			builder.WriteRune(r)
			w += runeWidth
		}
		builder.WriteString("...")
		return builder.String()
	}

	if currentWidth < targetLength {
		padding := targetLength - currentWidth
		builder.WriteString(s)
		builder.WriteString(strings.Repeat(" ", padding))
		return builder.String()
	}

	builder.WriteString(s)
	return builder.String()
}
