package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

type ProgramArgs struct {
	Debug             bool
	LogFile           string
	LogLevel          logrus.Level
	APIToken          string
	HostURL           string
	ListenPort        int
	CacheSize         int
	ModelInstanceURLs []string
}

func parseArgs() (ProgramArgs, error) {
	var args ProgramArgs
	var logLevel string
	var modelInstanceURLsStr string

	flag.BoolVar(&args.Debug, "debug", false, "enable debug logging for the bot")
	flag.StringVar(&args.LogFile, "log-file", "logs/bot.log", "file to write logs")
	flag.StringVar(&logLevel, "log-level", "info", "logging level for the bot (debug, info, warn, error)")
	flag.StringVar(&args.APIToken, "token", "", "Telegram Bot API token")
	flag.StringVar(&args.HostURL, "host-url", "", "your host machine URL")
	flag.IntVar(&args.ListenPort, "listen-port", 8080, "HTTP server port for webhook")
	flag.IntVar(&args.CacheSize, "cache-size", 1000, "size of the cache for processed files")
	flag.StringVar(&modelInstanceURLsStr, "model-instance-url", "", "Comma-separated list of model instance URLs in [url1,url2,...] format")

	flag.Parse()

	var err error
	args.LogLevel, err = logrus.ParseLevel(logLevel)
	if err != nil {
		return ProgramArgs{}, fmt.Errorf("error in parse log level: %v", err)
	}

	if args.APIToken == "" {
		return ProgramArgs{}, fmt.Errorf("error: -token is required")
	}
	if args.HostURL == "" {
		return ProgramArgs{}, fmt.Errorf("error: -host-url is required")
	}
	if modelInstanceURLsStr == "" {
		return ProgramArgs{}, fmt.Errorf("error: -model-instance-url is required")
	}

	modelInstanceURLsStr = strings.Trim(modelInstanceURLsStr, "[]")
	if modelInstanceURLsStr != "" {
		args.ModelInstanceURLs = strings.Split(modelInstanceURLsStr, ",")
		for i, url := range args.ModelInstanceURLs {
			args.ModelInstanceURLs[i] = strings.TrimSpace(url)
		}
	}

	return args, nil
}
