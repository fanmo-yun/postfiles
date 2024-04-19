package main

import "postfiles/flags"

func AppRun() {
	app := flags.Newflags()
	app.Parser()
	app.Handler()
	app.Run()
	println(app.Type, app.IP, app.Port)
}
