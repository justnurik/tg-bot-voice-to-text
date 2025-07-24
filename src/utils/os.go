package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func DownloadFile(url, fileName string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error in download file: [file url: %s] %v", url, err)
	}
	defer CloserErrorHandle(resp.Body, "error in closing response body")

	filePath := filepath.Join("downloads", fileName+".mp3")
	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("error in create file: [file name: %s] %v", fileName, err)
	}
	defer CloserErrorHandle(out, "error in closing output file")

	if _, err := io.Copy(out, resp.Body); err != nil {
		return "", fmt.Errorf("error in copy response body: %v", err)
	}

	return filePath, nil
}
