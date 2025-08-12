package stt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"tg-bot-voice-to-text/pkg/utils"
	"time"

	"go.uber.org/zap"
)

type STTClientDefault struct {
	logger *zap.Logger
}

func (s STTClientDefault) Request(filePath, url string) (string, error) {
	startTime := time.Now()
	log := s.logger.With(
		zap.String("worker_url", url),
		zap.String("file_path", filePath),
	)

	log.Info("STT request started")

	file, err := os.Open(filePath)
	if err != nil {
		log.Error("Error opening file", zap.Error(err), zap.String("file path", filePath))
		return "", fmt.Errorf("error opening file: %v", err)
	}
	defer utils.CloserErrorHandle(log, file, "Error closing file")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("audio", filepath.Base(filePath))
	if err != nil {
		log.Error("Error creating form file", zap.Error(err))
		return "", fmt.Errorf("error creating form file: %v", err)
	}

	bytesCopied, err := io.Copy(part, file)
	if err != nil {
		log.Error("Error copying file to form",
			zap.Error(err),
			zap.Int64("bytes_copied", bytesCopied))
		return "", fmt.Errorf("error copying file to form: %v", err)
	}

	if err := writer.Close(); err != nil {
		log.Error("Error closing multipart writer", zap.Error(err))
		return "", fmt.Errorf("error closing multipart writer: %v", err)
	}

	req, err := http.NewRequest("POST", url+"/transcriptions", body)
	if err != nil {
		log.Error("Error creating request", zap.Error(err))
		return "", fmt.Errorf("error creating new request: %v", err)
	}

	contentType := writer.FormDataContentType()
	req.Header.Set("Content-Type", contentType)
	defer utils.CloserErrorHandle(log, req.Body, "Error closing request body")

	log.Info("Request prepared",
		zap.String("content_type", contentType),
		zap.Int("body_size", body.Len()))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("Request failed",
			zap.Error(err),
			zap.Duration("elapsed", time.Since(startTime)))
		return "", fmt.Errorf("error performing request: %v", err)
	}
	defer utils.CloserErrorHandle(log, resp.Body, "Error closing response body")

	log.Info("Response received",
		zap.String("status", resp.Status),
		zap.Duration("elapsed", time.Since(startTime)))

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		log.Error("Unexpected status code",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response_body", string(bodyBytes)))
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Error reading response body",
			zap.Error(err),
			zap.Int("response_size", len(respData)))
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	log.Info("Response read",
		zap.Int("response_size", len(respData)))

	var response ResponseFromModel
	if err := json.Unmarshal(respData, &response); err != nil {
		log.Error("Error unmarshaling response",
			zap.Error(err),
			zap.String("response_sample", utils.Ellipsis(string(respData), 1024)))
		return "", fmt.Errorf("error unmarshaling response data: %v", err)
	}

	log.Info("STT request completed successfully",
		zap.String("transcription", utils.Ellipsis(response.Transcription, 1024)),
		zap.Duration("total_elapsed", time.Since(startTime)))

	return response.Transcription, nil
}

// REST API
type ResponseFromModel struct {
	Transcription string `json:"transcription"`
}
