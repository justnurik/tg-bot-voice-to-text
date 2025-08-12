package botwork

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type UpdateHandler interface {
	UpdateHandle(bot *tgbotapi.BotAPI, update *tgbotapi.Update) error
}
