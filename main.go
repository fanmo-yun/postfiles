package main

import "github.com/sirupsen/logrus"

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetLevel(logrus.InfoLevel)
	logrus.Info("app run")
	AppRun()
}
