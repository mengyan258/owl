package ocr

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type TencentConf struct {
	SecretID  string `json:"secret_id"`
	SecretKey string `json:"secret_key"`
	Region    string `json:"region"`
	Version   string `json:"version"`
}

type TencentClient struct{ conf TencentConf }

func NewTencentClient(c TencentConf) *TencentClient { return &TencentClient{conf: c} }

func (t *TencentClient) OCRText(imgBase64 string) (string, error) {
	// TC3-HMAC-SHA256 Signature
	service := "ocr"
	host := "ocr.tencentcloudapi.com"
	action := "GeneralBasicOCR"
	timestamp := time.Now().Unix()
	algorithm := "TC3-HMAC-SHA256"
	payload := map[string]any{"ImageBase64": imgBase64}
	body, _ := json.Marshal(payload)
	hashedPayload := sha256Hex(body)
	canonicalRequest := fmt.Sprintf("POST\n/\n\ncontent-type:application/json; charset=utf-8\nhost:%s\n\ncontent-type;host\n%s", host, hashedPayload)
	date := time.Unix(timestamp, 0).UTC().Format("2006-01-02")
	credentialScope := fmt.Sprintf("%s/%s/tc3_request", date, service)
	stringToSign := fmt.Sprintf("%s\n%d\n%s\n%s", algorithm, timestamp, credentialScope, sha256Hex([]byte(canonicalRequest)))
	secretDate := hmacSHA256([]byte("TC3"+t.conf.SecretKey), []byte(date))
	secretService := hmacSHA256(secretDate, []byte(service))
	secretSigning := hmacSHA256(secretService, []byte("tc3_request"))
	signature := hex.EncodeToString(hmacSHA256(secretSigning, []byte(stringToSign)))
	auth := fmt.Sprintf("%s Credential=%s/%s, SignedHeaders=content-type;host, Signature=%s", algorithm, t.conf.SecretID, credentialScope, signature)
	req, err := http.NewRequest("POST", "https://"+host, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Host", host)
	req.Header.Set("Authorization", auth)
	req.Header.Set("X-TC-Action", action)
	req.Header.Set("X-TC-Version", t.conf.Version)
	req.Header.Set("X-TC-Region", t.conf.Region)
	req.Header.Set("X-TC-Timestamp", strconv.FormatInt(timestamp, 10))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, resp.Body)
	var out struct {
		Response struct {
			TextDetections []struct {
				DetectedText string `json:"DetectedText"`
			} `json:"TextDetections"`
		} `json:"Response"`
	}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		return "", err
	}
	var s string
	for i, d := range out.Response.TextDetections {
		if i > 0 {
			s += " "
		}
		s += d.DetectedText
	}
	return s, nil
}

func sha256Hex(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func hmacSHA256(key []byte, msg []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(msg)
	return mac.Sum(nil)
}
