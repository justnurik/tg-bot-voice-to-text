package utils

import (
	"io"

	"github.com/sirupsen/logrus"
)

func CloserErrorHandle(c io.Closer, msg string) {
	if err := c.Close(); err != nil {
		logrus.Warnf("%s : %v", msg, err)
	}
}
