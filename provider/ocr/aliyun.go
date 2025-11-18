package ocr

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type AliyunConf struct {
	AppCode  string `json:"appcode"`
	Endpoint string `json:"endpoint"`
}

type AliyunClient struct{ conf AliyunConf }

func NewAliyunClient(c AliyunConf) *AliyunClient { return &AliyunClient{conf: c} }

func (a *AliyunClient) OCRText(imgBase64 string) (string, error) {
	payload := map[string]string{"img": imgBase64}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", a.conf.Endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "APPCODE "+a.conf.AppCode)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, resp.Body)
	var out struct {
		Data struct {
			Text string `json:"text"`
		} `json:"data"`
	}
	// 兼容不同市场返回结构，最佳化为直接取整个JSON中所有文本
	_ = json.Unmarshal(buf.Bytes(), &out)
	return out.Data.Text, nil
}
