package stt

import (
	"tg-bot-voice-to-text/pkg/scheduler"
	"tg-bot-voice-to-text/pkg/utils"
	"time"

	"go.uber.org/zap"
)

type STTServiceWithScheduler struct {
	logger *zap.Logger
	sched  scheduler.NamedWorkerScheduler[string]
	client STTClient
}

func NewSTTServiceWithScheduler(logger *zap.Logger, client STTClient, sched scheduler.NamedWorkerScheduler[string], instancesURL []string) STTServiceWithScheduler {
	logger.Info("Initializing STT service with scheduler",
		zap.Int("worker_count", len(instancesURL)),
		zap.Strings("worker_urls", instancesURL))

	s := STTServiceWithScheduler{
		logger: logger,
		sched:  sched,
		client: client,
	}
	s.sched.Start(instancesURL)
	return s
}

func (s STTServiceWithScheduler) TransformSpeechToText(filePath string) (string, error) {
	log := s.logger.With(zap.String("file_path", filePath))
	log.Info("Starting speech-to-text transformation")

	resultChan := make(chan string, 1)
	errChan := make(chan error, 1)

	startTime := time.Now()
	log.Info("Scheduling STT task")

	var workerURL string
	done := s.sched.Schedule(func(url string) {
		workerURL = url

		result, err := s.client.Request(filePath, url)

		resultChan <- result
		errChan <- err
	})

	<-done

	result := <-resultChan
	errResult := <-errChan

	if errResult != nil {
		log.Error("Speech-to-text transformation failed",
			zap.Error(errResult),
			zap.String("worker url", workerURL),
			zap.Duration("total_time", time.Since(startTime)))
	} else {
		log.Info("Speech-to-text transformation succeeded",
			zap.String("worker url", workerURL),
			zap.String("result_sample", utils.Ellipsis(result, 1024)),
			zap.Duration("total_time", time.Since(startTime)))
	}

	return result, errResult
}
