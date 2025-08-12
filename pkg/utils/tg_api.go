package utils

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func SendTextReply(bot *tgbotapi.BotAPI, chatID int64, replyToMessageID int, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyToMessageID = replyToMessageID
	if _, err := bot.Send(msg); err != nil {
		return fmt.Errorf("error in sending the text: [text: %s] %v", text, err)
	}
	return nil
}

func EditMessage(bot *tgbotapi.BotAPI, chatID int64, messageID int, text string) error {
	editMsg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	if _, err := bot.Send(editMsg); err != nil {
		return fmt.Errorf("error in editing message: [text: %s] %v", text, err)
	}
	return nil
}
