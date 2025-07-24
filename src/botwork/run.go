package botwork

import (
	"context"
	"fmt"
)

func RunOnWebHook(ctx context.Context, apiToken, hostURL, listenAddr string, uh UpdateHandler, debug bool) error {
	bot, err := NewWebHookBot(apiToken, uh, debug)
	if err != nil {
		return fmt.Errorf("error in bot init: %v", err)
	}

	if err := bot.Start(ctx, hostURL, listenAddr); err != nil {
		return fmt.Errorf("bot stopped: %v", err)
	}

	return nil
}

func RunOnLongPolling(ctx context.Context, apiToken, listenAddr string, uh UpdateHandler, timeout int, debug bool) error {
	bot, err := NewLongPollingBot(apiToken, uh, timeout, debug)
	if err != nil {
		return fmt.Errorf("error in bot init: %v", err)
	}

	if err := bot.Start(ctx); err != nil {
		return fmt.Errorf("bot stopped: %v", err)
	}

	return nil
}
