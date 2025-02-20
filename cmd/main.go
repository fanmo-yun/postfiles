package main

import (
	"postfiles/logger"
	"postfiles/utils"
)

func main() {
	logger.InitLogger()
	utils.IsTerminal()
	CliExecute()
}
