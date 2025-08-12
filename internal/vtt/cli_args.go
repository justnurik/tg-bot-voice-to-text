package vtt

import "flag"

type CLIArgs struct {
	BotConfigPath    string
	LoggerConfigPath string
}

func GetCLIArgs() *CLIArgs {
	var c CLIArgs

	flag.StringVar(&c.LoggerConfigPath, "logger-config-path", "./configs/logger.yml", "")
	flag.StringVar(&c.BotConfigPath, "bot-config-path", "./configs/bot.yml", "")

	flag.Parse()

	return &c
}
