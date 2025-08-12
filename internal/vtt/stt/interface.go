package stt

type STTClient interface {
	Request(filePath, url string) (string, error)
}

type STTService interface {
	TransformSpeechToText(voiceFilepath string) (string, error)
}
