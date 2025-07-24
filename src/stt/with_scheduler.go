package stt

import (
	"tg-bot-voice-to-text/src/scheduler"
)

type STTServiceWithScheduler struct {
	sched  scheduler.NamedWorkerScheduler[string]
	client STTClient
}

func NewSTTServiceWithScheduler(client STTClient, sched scheduler.NamedWorkerScheduler[string], instancesURL []string) STTServiceWithScheduler {
	s := STTServiceWithScheduler{
		sched:  sched,
		client: client,
	}
	s.sched.Start(instancesURL)
	return s
}

func (s STTServiceWithScheduler) TransformSpeechToText(filePath string) (string, error) {
	resultChan := make(chan string, 1)
	errChan := make(chan error, 1)

	done := s.sched.Schedule(func(url string) {
		result, err := s.client.Request(filePath, url)
		resultChan <- result
		errChan <- err
	})

	<-done

	return <-resultChan, <-errChan
}
