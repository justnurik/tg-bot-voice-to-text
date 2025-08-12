package botwork

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

func RunOnWebHook(ctx context.Context, logger *zap.Logger, name string, apiToken, listenAddr string, uh UpdateHandler, debug bool) error {
	logger.Info("initializing webhook bot",
		zap.String("listen_addr", listenAddr),
		zap.Bool("debug", debug),
	)

	bot, err := NewWebHookBot(logger, name, apiToken, uh, debug)
	if err != nil {
		logger.Error("failed to initialize webhook bot", zap.Error(err))
		return fmt.Errorf("error in bot init: %w", err)
	}

	logger.Info("starting webhook bot", zap.String("listen_addr", listenAddr))
	if err := bot.Start(ctx, listenAddr); err != nil {
		if ctx.Err() != nil {
			logger.Warn("webhook bot stopped due to context cancellation", zap.Error(ctx.Err()))
			return ctx.Err()
		}
		logger.Error("webhook bot stopped with error", zap.Error(err))
		return fmt.Errorf("bot stopped: %w", err)
	}

	logger.Info("webhook bot stopped normally")
	return nil
}

func RunOnLongPolling(ctx context.Context, logger *zap.Logger, name string, apiToken, listenAddr string, uh UpdateHandler, timeout int, debug bool) error {
	logger.Info("initializing long-polling bot",
		zap.String("listen_addr", listenAddr),
		zap.Int("timeout", timeout),
		zap.Bool("debug", debug),
	)

	bot, err := NewLongPollingBot(logger, name, apiToken, uh, timeout, debug)
	if err != nil {
		logger.Error("failed to initialize long-polling bot", zap.Error(err))
		return fmt.Errorf("error in bot init: %w", err)
	}

	logger.Info("starting long-polling bot")
	if err := bot.Start(ctx); err != nil {
		if ctx.Err() != nil {
			logger.Warn("long-polling bot stopped due to context cancellation", zap.Error(ctx.Err()))
			return ctx.Err()
		}
		logger.Error("long-polling bot stopped with error", zap.Error(err))
		return fmt.Errorf("bot stopped: %w", err)
	}

	logger.Info("long-polling bot stopped normally")
	return nil
}
