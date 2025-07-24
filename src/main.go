package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/sirupsen/logrus"

	"tg-bot-voice-to-text/src/botwork"
	"tg-bot-voice-to-text/src/queue"
	"tg-bot-voice-to-text/src/scheduler"
	"tg-bot-voice-to-text/src/stt"
	sttbot "tg-bot-voice-to-text/src/stt_bot"
)

func main() {
	// handle OS signals for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()

	// parse args
	programArgs, err := parseArgs()
	if err != nil {
		logrus.Fatalf("error in parse programm args: %v", err)
	}

	// logger setup
	logrus.SetLevel(programArgs.LogLevel)
	logFile, err := os.Open(programArgs.LogFile) // Ensure log file exists
	if err != nil {
		logrus.Fatalf("error in open log file: %v", err)
	}
	logrus.SetOutput(logFile)

	// requests cache
	cache, err := lru.New[string, string](programArgs.CacheSize)
	if err != nil {
		logrus.Fatalf("error in create cache: %v", err)
	}

	// load balancer(scheduler) for distributing requests across available ml instances of the voice-to-text translation model
	queue := queue.NewUnboundedChanQueue[func(string)]()
	sched := scheduler.NewNamedWorkerSchedulerQueue(ctx, queue)
	defer sched.Stop()

	// voice-to-text translation service
	stts := stt.NewSTTServiceWithScheduler(stt.STTClientDefault{}, sched, programArgs.ModelInstanceURLs)

	// strategy for getting a new update from the user
	STTBotUpdateHandler, err := sttbot.NewVoiceToTextUpdateHandler(stts, cache)
	if err != nil {
		logrus.Fatalf("error in create speech to text update handler: %v", err)
	}

	// create and run bot
	if err := botwork.RunOnWebHook(ctx, programArgs.APIToken, programArgs.HostURL, fmt.Sprintf(":%d", programArgs.ListenPort), STTBotUpdateHandler, programArgs.Debug); err != nil {
		logrus.Fatal(err)
	}
}
