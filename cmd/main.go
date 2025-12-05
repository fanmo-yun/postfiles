package main

import (
	"postfiles/cmdline"
	"postfiles/utils"
)

func main() {
	utils.IsTerm()
	cmdline.Execute()
}
