package setup

import (
	"errors"
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v3"
)

func Logger(configPath string) (*zap.Logger, error) {
	loggerConfigFile, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("error opening the logger configuration file: %v", err)
	}
	defer loggerConfigFile.Close()

	var loggerConfig LoggerConfig
	if err := yaml.NewDecoder(loggerConfigFile).Decode(&loggerConfig); err != nil {
		return nil, fmt.Errorf("error decoding the logger config: %v", err)
	}

	return loggerConfig.BuildLogger()
}

type LoggerConfig struct {
	AddStacktrace      bool   `yaml:"add-stacktrace"`
	StacktraceLogLevel string `yaml:"stacktrace-log-level"`
	AddCaller          bool   `yaml:"add-caller"`
	Console            bool   `yaml:"console"`
	ConsoleLogLevel    string `yaml:"console-log-level"`
	LogFilesConfig     []struct {
		FilePath   string `yaml:"file-path"`
		LogLevel   string `yaml:"log-level"`
		MaxSize    int    `yaml:"max-size"`
		MaxBackups int    `yaml:"max-backups"`
		MaxAge     int    `yaml:"max-age"`
	} `yaml:"log-files-config"`
}

func (config LoggerConfig) BuildLogger() (*zap.Logger, error) {
	var cores []zapcore.Core

	if config.Console {
		ConsoleLogLevel, err := stringToZapLogLevel(config.ConsoleLogLevel)
		if err != nil {
			return nil, err
		}

		consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		consoleWriter := zapcore.Lock(os.Stdout)
		consoleCore := zapcore.NewCore(consoleEncoder, consoleWriter, ConsoleLogLevel)
		cores = append(cores, consoleCore)
	}

	for _, fileConfig := range config.LogFilesConfig {
		lj := &lumberjack.Logger{
			Filename:   fileConfig.FilePath,
			MaxSize:    fileConfig.MaxSize,
			MaxBackups: fileConfig.MaxBackups,
			MaxAge:     fileConfig.MaxAge,
		}
		logLevel, err := stringToZapLogLevel(fileConfig.LogLevel)
		if err != nil {
			return nil, err
		}

		fileWriter := zapcore.AddSync(lj) // cast io.Writer -> zapcore.WriteSyncer
		fileEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
		fileCore := zapcore.NewCore(fileEncoder, fileWriter, logLevel)
		cores = append(cores, fileCore)
	}

	if len(cores) == 0 {
		return nil, errors.New("no logging outputs configured")
	}

	combinedCore := zapcore.NewTee(cores...)

	var opts []zap.Option
	if config.AddCaller {
		opts = append(opts, zap.AddCaller())
	}
	if config.AddStacktrace {
		stacktraceLogLevel, err := stringToZapLogLevel(config.StacktraceLogLevel)
		if err != nil {
			return nil, err
		}
		opts = append(opts, zap.AddStacktrace(stacktraceLogLevel))
	}

	logger := zap.New(combinedCore, opts...)
	return logger, nil
}

func stringToZapLogLevel(level string) (zapcore.Level, error) {
	switch level {
	case "debug", "DEBUG":
		return zap.DebugLevel, nil
	case "info", "INFO", "":
		return zap.InfoLevel, nil
	case "warn", "WARN":
		return zap.WarnLevel, nil
	case "error", "ERROR":
		return zap.ErrorLevel, nil
	case "dpanic", "DPANIC":
		return zap.DPanicLevel, nil
	case "panic", "PANIC":
		return zap.PanicLevel, nil
	case "fatal", "FATAL":
		return zap.FatalLevel, nil
	default:
		return zap.FatalLevel, fmt.Errorf("unsupported logging level")
	}
}
