package vtt

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	Token             string   `mapstructure:"token"`
	Mode              string   `mapstructure:"mode"` // webhook | longpoll
	Name              string   `mapstructure:"name"`
	Debug             bool     `mapstructure:"debug"`
	ListenAddr        string   `mapstructure:"listen_addr"`
	CacheSize         int      `mapstructure:"cache_size"`
	Timeout           int      `mapstructure:"timeout"` // for longpoll
	ModelInstanceURLs []string `mapstructure:"model_instance_urls"`
}

func LoadBotConfig(logger *zap.Logger, path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)

	v.SetEnvPrefix("BOT")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	_ = v.BindEnv("token")
	_ = v.BindEnv("mode")
	_ = v.BindEnv("name")
	_ = v.BindEnv("debug")
	_ = v.BindEnv("listen_addr")
	_ = v.BindEnv("cache_size")
	_ = v.BindEnv("timeout")
	_ = v.BindEnv("model_instance_urls")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.Info("config file not found, using environment variables only")
		} else {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if cfg.Token == "" {
		return nil, fmt.Errorf("bot token is required")
	}

	if cfg.Mode == "" {
		cfg.Mode = "longpoll"
	}
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = ":8080"
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 60
	}
	if cfg.CacheSize <= 0 {
		cfg.CacheSize = 100
	}

	logger.Info("loaded bot configuration",
		zap.String("mode", cfg.Mode),
		zap.String("name", cfg.Name),
		zap.String("listen_addr", cfg.ListenAddr),
		zap.Bool("debug", cfg.Debug),
		zap.Int("timeout", cfg.Timeout),
		zap.Int("cache_size", cfg.CacheSize),
		zap.Strings("model_instance_urls", cfg.ModelInstanceURLs),
	)

	return &cfg, nil
}
