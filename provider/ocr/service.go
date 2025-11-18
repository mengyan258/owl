package ocr

import (
	"encoding/base64"
	"errors"
	"io/ioutil"
)

type Client interface {
	OCRText(imgBase64 string) (string, error)
}

type Service struct {
	client Client
}

func NewService(c Client) *Service { return &Service{client: c} }

func (s *Service) OCRText(imgBytes []byte) (string, error) {
	b64 := base64.StdEncoding.EncodeToString(imgBytes)
	return s.client.OCRText(b64)
}

func (s *Service) OCRTextFromFile(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("empty image file")
	}
	return s.OCRText(data)
}
