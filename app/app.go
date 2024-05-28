package app

import (
	"postfiles/flags"
)

func parseAndHandle() {
	app := flags.Newflags()
	app.Parser()
	app.Handler()
	app.Run()
}

func Run() {
	parseAndHandle()
}
