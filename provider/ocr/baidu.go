package ocr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type BaiduConf struct {
	AK       string `json:"ak"`
	SK       string `json:"sk"`
	TokenURL string `json:"token_url"`
	OCRURL   string `json:"ocr_url"`
}

type BaiduClient struct{ conf BaiduConf }

func NewBaiduClient(conf BaiduConf) *BaiduClient { return &BaiduClient{conf: conf} }

func (b *BaiduClient) OCRText(imgBase64 string) (string, error) {
	token, err := b.getToken()
	if err != nil {
		return "", err
	}
	form := url.Values{}
	form.Set("image", imgBase64)
	req, err := http.NewRequest("POST", b.conf.OCRURL+"?access_token="+url.QueryEscape(token), bytes.NewBufferString(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, resp.Body)
	var out struct {
		WordsResult []struct {
			Words string `json:"words"`
		} `json:"words_result"`
	}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		return "", err
	}
	var s string
	for i, w := range out.WordsResult {
		if i > 0 {
			s += " "
		}
		s += w.Words
	}
	return s, nil
}

func (b *BaiduClient) getToken() (string, error) {
	u := fmt.Sprintf("%s?grant_type=client_credentials&client_id=%s&client_secret=%s", b.conf.TokenURL, url.QueryEscape(b.conf.AK), url.QueryEscape(b.conf.SK))
	resp, err := http.Get(u)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var out struct {
		AccessToken string `json:"access_token"`
	}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&out); err != nil {
		return "", err
	}
	if out.AccessToken == "" {
		return "", fmt.Errorf("baidu ocr: empty access token")
	}
	return out.AccessToken, nil
}
