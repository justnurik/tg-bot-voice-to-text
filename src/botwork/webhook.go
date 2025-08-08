package botwork

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"tg-bot-voice-to-text/src/utils"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

type WebHookBot struct {
	bot *tgbotapi.BotAPI
	uh  UpdateHandler
}

func NewWebHookBot(apiToken string, uh UpdateHandler, debug bool) (*WebHookBot, error) {
	bot, err := tgbotapi.NewBotAPI(apiToken)
	if err != nil {
		return nil, err
	}

	bot.Debug = debug

	return &WebHookBot{
		bot: bot,
		uh:  uh,
	}, nil
}

func (w WebHookBot) Start(ctx context.Context, hostURL, listenAddr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", w.webhookHandler)

	server := &http.Server{
		Addr:    listenAddr,
		Handler: mux,
	}

	errChan := make(chan error, 1)
	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("error starting HTTPS server: %v", err)
		}
	}()

	<-time.After(1 * time.Second) // Wait for server to start

	if err := w.setWebhook(fmt.Sprintf("https://%s/webhook", hostURL)); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
	case err := <-errChan:
		return err
	}

	return server.Shutdown(context.Background())
}

// TODO: err chan
func (e WebHookBot) webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var update tgbotapi.Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, fmt.Errorf("error decoding update: %v", err).Error(), http.StatusBadRequest)
		return
	}
	defer utils.CloserErrorHandle(r.Body, "error closing body")

	//
	go func() {
		if err := e.uh.UpdateHandle(e.bot, &update); err != nil {
			logrus.Errorf("error in one update handler: %v", err)
		}
	}()
}

func (e WebHookBot) setWebhook(webhookURL string) error {
	webhook, err := tgbotapi.NewWebhook(webhookURL)
	if err != nil {
		return fmt.Errorf("error creating webhook: %v", err)
	}

	if _, err := e.bot.Request(webhook); err != nil {
		return fmt.Errorf("error setting webhook: %v", err)
	}

	info, err := e.bot.GetWebhookInfo()
	if err != nil {
		return fmt.Errorf("error getting webhook info: %v", err)
	}

	if info.LastErrorDate != 0 {
		return fmt.Errorf("webhook last error: %s", info.LastErrorMessage)
	}

	return nil
}
