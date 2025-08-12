package utils

import (
	"io"

	"go.uber.org/zap"
)

func CloserErrorHandle(logger *zap.Logger, c io.Closer, msg string) {
	if err := c.Close(); err != nil {
		logger.Debug(msg, zap.Error(err))
	}
}
