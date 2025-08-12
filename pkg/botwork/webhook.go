package botwork

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"tg-bot-voice-to-text/pkg/utils"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	bodyLogSize = 100
)

type WebHookBot struct {
	logger *zap.Logger

	bot *tgbotapi.BotAPI
	uh  UpdateHandler
}

func NewWebHookBot(logger *zap.Logger, name string, apiToken string, uh UpdateHandler, debug bool) (*WebHookBot, error) {
	bot, err := tgbotapi.NewBotAPI(apiToken)
	if err != nil {
		return nil, err
	}

	bot.Debug = debug

	logger = logger.With(zap.String("name", name))
	logger.Info("create new webhook bot")

	return &WebHookBot{
		logger: logger,
		bot:    bot,
		uh:     uh,
	}, nil
}

func (w *WebHookBot) Start(ctx context.Context, listenAddr string) error {
	mux := http.NewServeMux()
	mux.Handle("/webhook/bot2", w.loggingMiddleware(http.HandlerFunc(w.newWebhookHandler())))

	server := &http.Server{
		Addr:    listenAddr,
		Handler: mux,
	}

	errChan := make(chan error, 1)
	go func() {
		w.logger.Info("start server for webhooks")

		if err := server.ListenAndServe(); err != nil {
			w.logger.Error("error starting HTTP server", zap.Error(err))
			errChan <- fmt.Errorf("error starting HTTP server: %v", err)
		}
	}()

	select {
	case <-ctx.Done():
		w.logger.Warn("contex done (timeout/context canceled/...)", zap.Error(ctx.Err()))
	case err := <-errChan:
		return err
	}

	w.logger.Info("try shutdown server")

	defer w.logger.Info("shutdown server")
	return server.Shutdown(context.Background())
}

func (e *WebHookBot) newWebhookHandler() func(http.ResponseWriter, *http.Request) {
	logger := e.logger.With(zap.String("component", "webhook handler"))

	handler := func(w http.ResponseWriter, r *http.Request) {
		logger := logger.With(zap.String("id", uuid.New().String()))

		if r.Method != http.MethodPost {
			logger.Error("method not allowed, only POST", zap.String("method", r.Method))
			http.Error(w, "Method not allowed, only POST", http.StatusMethodNotAllowed)
			return
		}

		var update tgbotapi.Update
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			body, err := io.ReadAll(io.LimitReader(r.Body, bodyLogSize))
			if err != nil {
				logger.Error("error decoding update", zap.Error(err),
					zap.String("body", "failed read body"))
				return
			}

			logger.Error("error decoding update", zap.Error(err),
				zap.String("body", string(body)))
			http.Error(w, fmt.Errorf("error decoding update: %v", err).Error(), http.StatusBadRequest)
			return
		}
		defer utils.CloserErrorHandle(logger, r.Body, "error closing body")

		logger.Info("start update handler")
		if err := e.uh.UpdateHandle(e.bot, &update); err != nil {
			logger.Error("error in one update handler", zap.Error(err))
			http.Error(w, fmt.Errorf("error in one update handler: %v", err).Error(), http.StatusInternalServerError)
		}
		logger.Info("finish update handler")
	}

	return handler
}

func (w WebHookBot) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		start := time.Now()

		bodyBytes, _ := io.ReadAll(io.LimitReader(r.Body, bodyLogSize))
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		w.logger.Info("incoming webhook request",
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
			zap.Int("content_length", len(bodyBytes)),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("user_agent", r.UserAgent()),
			zap.ByteString("body", bodyBytes),
		)

		next.ServeHTTP(rw, r)

		w.logger.Info("webhook handled",
			zap.Duration("duration", time.Since(start)),
		)
	})
}
