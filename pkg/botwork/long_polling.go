package botwork

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type LongPollingBot struct {
	name   string
	logger *zap.Logger

	bot     *tgbotapi.BotAPI
	updates tgbotapi.UpdatesChannel
	uh      UpdateHandler
}

func NewLongPollingBot(logger *zap.Logger, name string, apiToken string, uh UpdateHandler, timeout int, debug bool) (*LongPollingBot, error) {
	bot, err := tgbotapi.NewBotAPI(apiToken)
	if err != nil {
		return nil, err
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = timeout
	bot.Debug = debug

	logger = logger.With(zap.String("name", name))
	logger.Info("create new long-polling bot")

	return &LongPollingBot{
		name:    name,
		logger:  logger,
		bot:     bot,
		updates: bot.GetUpdatesChan(u),
		uh:      uh,
	}, nil
}

func (lpb *LongPollingBot) Start(ctx context.Context) error {
	lpb.logger.Info("start bot")

	for {
		select {

		case <-ctx.Done():
			lpb.logger.Error("contex done (timeout/context canceled/...)", zap.Error(ctx.Err()))
			return ctx.Err()

		case update := <-lpb.updates:
			lpb.logger.Info("start update handle")

			if err := lpb.uh.UpdateHandle(lpb.bot, &update); err != nil {
				lpb.logger.Error("failed update handle", zap.Error(err))
				return fmt.Errorf("error in update handler: %v", err)
			}

			lpb.logger.Info("done update handle")
		}
	}
}
