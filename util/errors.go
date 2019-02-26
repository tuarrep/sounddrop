package util

import (
	"github.com/sirupsen/logrus"
)

func CheckError(err error, logger *logrus.Entry) {
	if err != nil {
		logger.Error(err)
		panic(err)
	}
}
