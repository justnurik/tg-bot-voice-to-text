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
	"tg-bot-voice-to-text/src/utils"

	"github.com/sirupsen/logrus"
)

type STTClientDefault struct{}

func (s STTClientDefault) Request(filePath, url string) (string, error) {
	logrus.Infof("start [worker %s] [file %s]", url, filePath)

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error opening file: %v", err)
	}
	defer utils.CloserErrorHandle(file, "error closing file")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("audio", filepath.Base(filePath))
	if err != nil {
		return "", fmt.Errorf("error creating form file: %v", err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return "", fmt.Errorf("error copying file to form: %v", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("error closing multipart writer: %v", err)
	}

	req, err := http.NewRequest("POST", url+"/transcriptions", body)
	if err != nil {
		return "", fmt.Errorf("error creating new request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	defer utils.CloserErrorHandle(req.Body, "error closing request body")

	logrus.Infof("request created [worker %s] [file %s]", url, filePath)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error performing request: %v", err)
	}
	defer utils.CloserErrorHandle(resp.Body, "error closing response body")

	logrus.Infof("client::Do [worker %s] [file %s]", url, filePath)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	logrus.Infof("read response [worker %s] [file %s]", url, filePath)

	var response ResponseFromModel
	if err := json.Unmarshal(respData, &response); err != nil {
		return "", fmt.Errorf("error unmarshaling response data: %v", err)
	}

	logrus.Infof("response unmarshaled [worker %s] [file %s]", url, filePath)

	logrus.Infof("result sent [worker %s] [file %s] [result: %s]", url, filePath, response.Transcription)

	return response.Transcription, nil
}

// REST API
type ResponseFromModel struct {
	Transcription string `json:"transcription"`
}
