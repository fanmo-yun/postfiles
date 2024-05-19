package app

import (
	"fmt"
	"os"
	"path"
	"postfiles/flags"
	"time"

	"github.com/sirupsen/logrus"
)

func configureLogger() *os.File {
	logFilename := fmt.Sprintf("%d.log", time.Now().Unix())
	logDir := "log"

	if mkdirallErr := os.MkdirAll(logDir, os.ModePerm); mkdirallErr != nil {
		fmt.Println("Error creating log directory:", mkdirallErr)
	}

	fp, openfileErr := os.OpenFile(path.Join(logDir, logFilename), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if openfileErr != nil {
		fmt.Println("Error opening log file:", openfileErr)
	}

	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetOutput(fp)

	return fp
}

func parseAndHandleFlags() {
	app := flags.Newflags()
	app.Parser()
	app.Handler()
	app.Run()
}

func Run() {
	defer configureLogger().Close()
	parseAndHandleFlags()
}
