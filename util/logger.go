package util

import (
	"github.com/sirupsen/logrus"
	"os"
)

func InitLogger() {
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)
}

func GetContextLogger(filename string, unitName string) *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"file": filename,
		"unit": unitName,
	})
}
