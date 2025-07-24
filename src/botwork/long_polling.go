package botwork

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type LongPollingBot struct {
	bot     *tgbotapi.BotAPI
	updates tgbotapi.UpdatesChannel
	uh      UpdateHandler
}

func NewLongPollingBot(apiToken string, uh UpdateHandler, timeout int, debug bool) (*LongPollingBot, error) {
	bot, err := tgbotapi.NewBotAPI(apiToken)
	if err != nil {
		return nil, err
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = timeout
	bot.Debug = debug

	return &LongPollingBot{
		bot:     bot,
		updates: bot.GetUpdatesChan(u),
		uh:      uh,
	}, nil
}

func (lpb *LongPollingBot) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case update := <-lpb.updates:
			if err := lpb.uh.UpdateHandle(lpb.bot, &update); err != nil {
				return fmt.Errorf("error in update handler: %v", err)
			}
		}
	}
}
