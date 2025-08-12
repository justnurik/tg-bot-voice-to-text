package utils

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

func ReadConfig(logger *zap.Logger, path string, configStruct any) error {
	f, err := os.Open(path)
	if err != nil {
		logger.Error("couldn't open the config file", zap.Error(err))
		return fmt.Errorf("couldn't open the config file: %v", err)
	}
	defer CloserErrorHandle(logger, f, "failed file close")

	if err := yaml.NewDecoder(f).Decode(configStruct); err != nil {
		logger.Error("failed decode config", zap.Error(err))
		return fmt.Errorf("failed decode config: %v", err)
	}
	logger.Info("read config", zap.String("config path", path))

	return nil
}
