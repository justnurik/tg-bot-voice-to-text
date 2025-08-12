package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"tg-bot-voice-to-text/internal/vtt"
	"tg-bot-voice-to-text/internal/vtt/stt"
	"tg-bot-voice-to-text/pkg/botwork"
	"tg-bot-voice-to-text/pkg/cache"
	"tg-bot-voice-to-text/pkg/queue"
	"tg-bot-voice-to-text/pkg/scheduler"
	"tg-bot-voice-to-text/pkg/setup"

	lru "github.com/hashicorp/golang-lru/v2"
	"go.uber.org/zap"
)

func main() {
	// get CLI args [bot|logger]-config-path
	args := vtt.GetCLIArgs()

	// logger setup
	logger, err := setup.Logger(args.LoggerConfigPath)
	if err != nil {
		panic(fmt.Errorf("failed logger setup: %v", err))
	}
	defer logger.Sync()

	// load bot config
	cfg, err := vtt.LoadBotConfig(logger, args.BotConfigPath)
	if err != nil {
		logger.Fatal("failed read bot config", zap.Error(err))
	}

	// context + graceful shutdown
	logger.Info("Setting up graceful shutdown")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		logger.Info("signal received, shutting down")
		cancel()
	}()

	// Initialize components
	logger.Info("Initializing components")
	logger.Debug("Creating task queue")
	queue := queue.NewUnboundedChanQueue[func(string)]()
	logger.Debug("Creating worker scheduler")
	sched := scheduler.NewNamedWorkerSchedulerQueue(ctx, queue)

	logger.Info("Setting up file ID cache", zap.Int("size", cfg.CacheSize))
	var fileIDCache cache.Cache[string, string] = nil
	fileIDCache, err = lru.New[string, string](cfg.CacheSize)
	if err != nil {
		logger.Error("Failed to create LRU cache, using no-op cache",
			zap.Error(err),
			zap.Int("cache_size", cfg.CacheSize))
		fileIDCache = cache.EmptyCache[string, string]{}
	} else {
		logger.Info("LRU cache created successfully")
	}

	logger.Info("Initializing STT service",
		zap.Int("worker_count", len(cfg.ModelInstanceURLs)))
	sttService := stt.NewSTTServiceWithScheduler(logger, stt.STTClientDefault{}, sched, cfg.ModelInstanceURLs)

	logger.Info("Creating update handler")
	uh, err := vtt.NewVoiceToTextUpdateHandler(logger, sttService, fileIDCache)
	if err != nil {
		logger.Fatal("Failed to create update handler", zap.Error(err))
	}

	// Start bot
	switch cfg.Mode {
	case "webhook":
		logger.Info("starting webhook mode", zap.String("listen_addr", cfg.ListenAddr))
		if err := botwork.RunOnWebHook(
			ctx, logger, cfg.Name,
			cfg.Token, cfg.ListenAddr,
			uh, cfg.Debug,
		); err != nil {
			logger.Error("webhook stopped", zap.Error(err))
		}

	case "longpoll":
		logger.Info("starting longpoll mode", zap.String("listen_addr", cfg.ListenAddr), zap.Int("timeout", cfg.Timeout))
		if err := botwork.RunOnLongPolling(
			ctx, logger, cfg.Name,
			cfg.Token, cfg.ListenAddr,
			uh, cfg.Timeout, cfg.Debug,
		); err != nil {
			logger.Error("longpoll stopped", zap.Error(err))
		}

	default:
		logger.Fatal("unknown mode, must be webhook or longpoll", zap.String("mode", cfg.Mode))
	}

	logger.Info("Application exited gracefully")
}
