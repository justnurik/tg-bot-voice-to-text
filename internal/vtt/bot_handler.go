package vtt

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"tg-bot-voice-to-text/internal/vtt/stt"
	"tg-bot-voice-to-text/pkg/cache"
	"tg-bot-voice-to-text/pkg/utils"
)

const (
	voice = iota
	audio
	videoNote
	skipMessage
)

type SpeechToTextUpdateHandler struct {
	logger             *zap.Logger
	stts               stt.STTService
	processedFileCache cache.Cache[string, string]
}

func NewVoiceToTextUpdateHandler(logger *zap.Logger, stts stt.STTService, cache cache.Cache[string, string]) (*SpeechToTextUpdateHandler, error) {
	logger = logger.Named("vtt-handler")

	if err := os.Mkdir("./downloads", 0755); !errors.Is(err, os.ErrExist) && err != nil {
		logger.Error("Failed to create downloads directory", zap.Error(err))
		return nil, fmt.Errorf("error in create directory 'downloads': %v", err)
	}

	logger.Info("Handler initialized", zap.String("downloads_dir", "./downloads"))
	return &SpeechToTextUpdateHandler{
		logger:             logger,
		stts:               stts,
		processedFileCache: cache,
	}, nil
}

func (v *SpeechToTextUpdateHandler) UpdateHandle(bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
	if update.Message == nil {
		return nil
	}

	log := v.logger.With(
		zap.Int64("chat_id", update.Message.Chat.ID),
		zap.Int("message_id", update.Message.MessageID),
		zap.String("user", update.Message.From.UserName),
	)
	log.Info("Processing new message")

	fileID, msgText, state := v.chooseReactionOnMessage(update.Message)
	log = log.With(zap.String("file_id", fileID), zap.Int("media_type", state))

	sentMsg, err := v.ReactionOnMessage(bot, update.Message, fileID, msgText, state)
	if err != nil {
		log.Error("Reaction on message failed", zap.Error(err))
		return fmt.Errorf("error in reaction on message: %v", err)
	}
	if sentMsg == nil { // skip message
		log.Info("Message skipped (not media)")
		return nil
	}

	cacheHit, err := v.cacheHitCheck(bot, update.Message, sentMsg, fileID)
	if err != nil {
		log.Error("Cache check failed", zap.Error(err))
		return fmt.Errorf("error in cache hit check: %v", err)
	}
	if cacheHit {
		log.Info("Cache hit, returning cached result")
		return nil
	}

	filepath, err := v.downloadFile(bot, update.Message, sentMsg, fileID)
	if err != nil {
		log.Error("File download failed", zap.Error(err))
		return fmt.Errorf("error in get file for transcription: %v", err)
	}
	if filepath == "" {
		return nil
	}
	defer func() {
		if err := os.Remove(filepath); err != nil {
			log.Warn("Failed to remove temp file",
				zap.String("file_path", filepath),
				zap.Error(err))
		} else {
			log.Debug("Temp file removed", zap.String("file_path", filepath))
		}
	}()

	log = log.With(zap.String("file_path", filepath))
	log.Info("File downloaded successfully")

	transcription, err := v.transcription(bot, update.Message, sentMsg, filepath)
	if err != nil {
		log.Error("Transcription failed", zap.Error(err))
		return fmt.Errorf("error in transcription: %v", err)
	}
	if transcription == "" {
		log.Info("Empty transcription result")
		return nil
	}

	if err := utils.EditMessage(bot, update.Message.Chat.ID, sentMsg.MessageID, transcription); err != nil {
		log.Error("Failed to edit message",
			zap.String("transcription", transcription),
			zap.Error(err))
		return fmt.Errorf("error in edit message: [text: %s] %v", transcription, err)
	}

	v.processedFileCache.Add(fileID, transcription)
	log.Info("Transcription completed",
		zap.String("transcription", utils.Ellipsis(transcription, 50)))

	return nil
}

func (v SpeechToTextUpdateHandler) chooseReactionOnMessage(message *tgbotapi.Message) (string, string, int) {
	switch {

	case message.Audio != nil:
		return message.Audio.FileID, "Получено аудио, обрабатываю...", audio

	case message.Voice != nil:
		return message.Voice.FileID, "Получено голосовое сообщение, обрабатываю...", voice

	case message.VideoNote != nil:
		return message.VideoNote.FileID, "Получено видео сообщение, обрабатываю...", videoNote

	default:
		return "", "Отправте голосовое сообщение!", skipMessage
	}
}

func (v SpeechToTextUpdateHandler) ReactionOnMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message, fileID, msgText string, state int) (*tgbotapi.Message, error) {
	// reaction: skip message
	if state == skipMessage {
		if message.Chat.Type == "private" {
			if err := utils.SendTextReply(bot, message.Chat.ID, message.MessageID, msgText); err != nil {
				return nil, fmt.Errorf("error in send text: [text: %s] %v", msgText, err)
			}
		}

		return nil, nil // skip message
	}

	// reaction: notification to the user about the start of processing
	acceptedMsg := tgbotapi.NewMessage(message.Chat.ID, msgText)
	acceptedMsg.ReplyToMessageID = message.MessageID
	sentMsg, err := bot.Send(acceptedMsg)
	if err != nil {
		return nil, fmt.Errorf("error in sending accepted message: %v", err)
	}

	return &sentMsg, nil
}

func (v SpeechToTextUpdateHandler) cacheHitCheck(bot *tgbotapi.BotAPI, message, sentMsg *tgbotapi.Message, fileID string) (bool, error) {
	// check cache
	if text, exist := v.processedFileCache.Get(fileID); exist {
		if err := utils.EditMessage(bot, message.Chat.ID, sentMsg.MessageID, text); err != nil {
			return false, fmt.Errorf("error in send message: [text: %s] %v", text, err)
		}

		return true, nil // cache hit
	}
	// cache miss

	return false, nil
}

func (v SpeechToTextUpdateHandler) downloadFile(bot *tgbotapi.BotAPI, message, sentMsg *tgbotapi.Message, fileID string) (string, error) {
	fileURL, err := bot.GetFileDirectURL(fileID)
	if err != nil {
		logrus.Errorf("error in get file direct url: [file id: %s] %v", fileID, err)
		if err := utils.EditMessage(bot, message.Chat.ID, message.MessageID, "Ошибка получения файла"); err != nil {
			return "", fmt.Errorf("error in edit message: %v", err)
		}
		return "", nil
	}

	filePath, err := utils.DownloadFile(v.logger, fileURL, fmt.Sprintf("tmp_%s", uuid.New()))
	if err != nil {
		logrus.Errorf("error in download file: [file url: %s] %v", fileURL, err)
		if err := utils.EditMessage(bot, message.Chat.ID, sentMsg.MessageID, "Ошибка скачивания файла"); err != nil {
			return "", fmt.Errorf("error in edit message: %v", err)
		}
		return "", nil
	}

	absFilepath, err := filepath.Abs(filePath)
	if err != nil {
		_ = os.Remove(filePath)
		return "", fmt.Errorf("error in transform path to abs path: %v", err)
	}

	return absFilepath, nil
}

func (v SpeechToTextUpdateHandler) transcription(bot *tgbotapi.BotAPI, message, sentMsg *tgbotapi.Message, filepath string) (string, error) {
	transcription, err := v.stts.TransformSpeechToText(filepath)
	if err != nil {
		logrus.Errorf("error in transcription: [file path: %s] %v", filepath, err)
		if err := utils.EditMessage(bot, message.Chat.ID, sentMsg.MessageID, "Ошибка транскрипции в текст :("); err != nil {
			return "", fmt.Errorf("error in edit message: %v", err)
		}
		return "", nil
	}

	if transcription == "" {
		transcription = "Текста в аудио нету."
	}

	return transcription, nil
}
