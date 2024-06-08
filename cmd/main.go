package main

import (
	"postfiles/app"
	"postfiles/utils"
)

func main() {
	utils.IsTerminal()
	app.Run()
}
