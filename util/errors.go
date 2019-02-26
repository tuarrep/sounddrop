package util

import (
	"github.com/sirupsen/logrus"
)

// CheckError check if error is not nil and panic
func CheckError(err error, logger *logrus.Entry) {
	if err != nil {
		logger.Error(err)
		panic(err)
	}
}
