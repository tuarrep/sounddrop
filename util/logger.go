package util

import (
	"github.com/sirupsen/logrus"
	"os"
)

// InitLogger initialize logger service
func InitLogger() {
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)
}

// GetContextLogger get a logger instance contextualized by service name
func GetContextLogger(filename string, unitName string) *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"file": filename,
		"unit": unitName,
	})
}
