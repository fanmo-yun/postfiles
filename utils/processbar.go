package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/term"
)

func NewBar(size int64, name string) (*progressbar.ProgressBar, error) {
	desc, err := fitText(name)
	if err != nil {
		return nil, err
	}
	bar := progressbar.NewOptions64(
		size,
		progressbar.OptionSetDescription(desc),
		progressbar.OptionShowBytes(true),
		progressbar.OptionFullWidth(),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprintln(os.Stdout)
		}),
	)
	return bar, nil
}

func termWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80
	}
	return w
}

func textWidth(s string) int {
	w := 0
	for _, r := range s {
		w += runewidth.RuneWidth(r)
	}
	return w
}

func fitText(s string) (string, error) {
	w := termWidth()
	target := w * 30 / 100

	return clipOrPad(s, target)
}

func clipOrPad(s string, limit int) (string, error) {
	cur := textWidth(s)

	if cur > limit {
		var b strings.Builder
		w := 0
		for _, r := range s {
			rw := runewidth.RuneWidth(r)
			if w+rw > limit-3 {
				break
			}
			if _, err := b.WriteRune(r); err != nil {
				return "", err
			}
			w += rw
		}
		_, err := b.WriteString("...")
		return b.String(), err
	}

	if cur < limit {
		return s + strings.Repeat(" ", limit-cur), nil
	}

	return s, nil
}
