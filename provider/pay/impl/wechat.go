package impl

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	mathrand "math/rand"
	"net/http"
	"strings"
	"time"

	payc "bit-labs.cn/owl/contract/pay"
)

type Wechat struct {
	cfg *payc.WechatConfig
}

func NewWechat(cfg *payc.WechatConfig) (*Wechat, error) {
	return &Wechat{cfg: cfg}, nil
}

func (w *Wechat) Create(ctx context.Context, intent *payc.PayIntent) (*payc.ClientAction, *payc.Transaction, error) {
	if intent == nil {
		return nil, nil, &payc.PaymentError{Code: "Invalid", Message: "intent nil"}
	}
	if intent.Scene == payc.SceneNative {
		urlStr := "https://api.mch.weixin.qq.com/v3/pay/transactions/native"
		body := map[string]any{
			"mchid":        w.cfg.MchID,
			"appid":        w.cfg.AppID,
			"description":  intent.Subject,
			"out_trade_no": intent.OutTradeNo,
			"notify_url":   w.cfg.NotifyURL,
			"amount":       map[string]any{"total": intent.Amount},
		}
		b, _ := json.Marshal(body)
		req, err := http.NewRequestWithContext(ctx, "POST", urlStr, strings.NewReader(string(b)))
		if err != nil {
			return nil, nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		auth, err := w.sign("POST", "/v3/pay/transactions/native", string(b))
		if err != nil {
			return nil, nil, err
		}
		req.Header.Set("Authorization", auth)
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return nil, nil, err
		}
		defer resp.Body.Close()
		rb, _ := ioutil.ReadAll(resp.Body)
		var r map[string]any
		_ = json.Unmarshal(rb, &r)
		codeURL := ""
		if v, ok := r["code_url"].(string); ok {
			codeURL = v
		}
		ca := &payc.ClientAction{Kind: payc.ActionQRCode, Data: map[string]string{"code_url": codeURL}}
		tx := &payc.Transaction{Status: payc.TxCreated, Provider: "wechat", Channel: payc.ChannelWechat, Scene: intent.Scene, OutTradeNo: intent.OutTradeNo, Amount: intent.Amount, Currency: intent.Currency, Payer: intent.Payer, Raw: r}
		return ca, tx, nil
	}
	if intent.Scene == payc.SceneJSAPI || intent.Scene == payc.SceneMiniProgram {
		urlStr := "https://api.mch.weixin.qq.com/v3/pay/transactions/jsapi"
		body := map[string]any{
			"mchid":        w.cfg.MchID,
			"appid":        w.cfg.AppID,
			"description":  intent.Subject,
			"out_trade_no": intent.OutTradeNo,
			"notify_url":   w.cfg.NotifyURL,
			"amount":       map[string]any{"total": intent.Amount},
			"payer":        map[string]any{"openid": intent.Payer.OpenID},
		}
		b, _ := json.Marshal(body)
		req, err := http.NewRequestWithContext(ctx, "POST", urlStr, strings.NewReader(string(b)))
		if err != nil {
			return nil, nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		auth, err := w.sign("POST", "/v3/pay/transactions/jsapi", string(b))
		if err != nil {
			return nil, nil, err
		}
		req.Header.Set("Authorization", auth)
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return nil, nil, err
		}
		defer resp.Body.Close()
		rb, _ := ioutil.ReadAll(resp.Body)
		var r map[string]any
		_ = json.Unmarshal(rb, &r)
		prepayId := ""
		if v, ok := r["prepay_id"].(string); ok {
			prepayId = v
		}
		ts := fmt.Sprintf("%d", time.Now().Unix())
		nonce := randStr(16)
		pkg := "prepay_id=" + prepayId
		paySign, err := w.paySign(ts, nonce, pkg)
		if err != nil {
			return nil, nil, err
		}
		data := map[string]string{
			"appId":     w.cfg.AppID,
			"timeStamp": ts,
			"nonceStr":  nonce,
			"package":   pkg,
			"signType":  "RSA",
			"paySign":   paySign,
		}
		ca := &payc.ClientAction{Kind: payc.ActionMiniParams, Data: data}
		tx := &payc.Transaction{Status: payc.TxCreated, Provider: "wechat", Channel: payc.ChannelWechat, Scene: intent.Scene, OutTradeNo: intent.OutTradeNo, Amount: intent.Amount, Currency: intent.Currency, Payer: intent.Payer, Raw: r}
		return ca, tx, nil
	}
	return nil, nil, &payc.PaymentError{Code: "UnsupportedScene", Message: string(intent.Scene)}
}

func (w *Wechat) Query(ctx context.Context, req *payc.QueryRequest) (*payc.Transaction, error) {
	if req.OutTradeNo == "" && req.ProviderTxnId == "" {
		return nil, &payc.PaymentError{Code: "Invalid", Message: "missing identifiers"}
	}
	path := ""
	if req.OutTradeNo != "" {
		path = "/v3/pay/transactions/out-trade-no/" + req.OutTradeNo + "?mchid=" + w.cfg.MchID
	} else {
		path = "/v3/pay/transactions/id/" + req.ProviderTxnId + "?mchid=" + w.cfg.MchID
	}
	urlStr := "https://api.mch.weixin.qq.com" + path
	reqh, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	auth, err := w.sign("GET", path, "")
	if err != nil {
		return nil, err
	}
	reqh.Header.Set("Authorization", auth)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(reqh)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	rb, _ := ioutil.ReadAll(resp.Body)
	var r map[string]any
	_ = json.Unmarshal(rb, &r)
	st := payc.TxPending
	if v, ok := r["trade_state"].(string); ok {
		switch v {
		case "SUCCESS":
			st = payc.TxPaid
		case "NOTPAY", "USERPAYING":
			st = payc.TxPending
		case "CLOSED", "PAYERROR":
			st = payc.TxFailed
		default:
			st = payc.TxPending
		}
	}
	tx := &payc.Transaction{Status: st, Provider: "wechat", OutTradeNo: req.OutTradeNo, ProviderTxnId: fmt.Sprintf("%v", r["transaction_id"]), Raw: r}
	return tx, nil
}

func (w *Wechat) Close(ctx context.Context, req *payc.CloseRequest) error {
	path := "/v3/pay/transactions/out-trade-no/" + req.OutTradeNo + "/close"
	urlStr := "https://api.mch.weixin.qq.com" + path
	body := map[string]any{"mchid": w.cfg.MchID}
	b, _ := json.Marshal(body)
	reqh, err := http.NewRequestWithContext(ctx, "POST", urlStr, strings.NewReader(string(b)))
	if err != nil {
		return err
	}
	reqh.Header.Set("Content-Type", "application/json")
	auth, err := w.sign("POST", path, string(b))
	if err != nil {
		return err
	}
	reqh.Header.Set("Authorization", auth)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(reqh)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 204 && resp.StatusCode != 200 {
		rb, _ := ioutil.ReadAll(resp.Body)
		return &payc.PaymentError{Code: fmt.Sprintf("%d", resp.StatusCode), Message: string(rb)}
	}
	return nil
}

func (w *Wechat) Refund(ctx context.Context, req *payc.RefundRequest) (*payc.RefundResult, error) {
	urlStr := "https://api.mch.weixin.qq.com/v3/refund/domestic/refunds"
	body := map[string]any{
		"out_refund_no": req.OutRefundNo,
		"amount":        map[string]any{"refund": req.Amount, "total": req.Amount, "currency": "CNY"},
	}
	if req.OutTradeNo != "" {
		body["out_trade_no"] = req.OutTradeNo
	}
	if req.ProviderTxnId != "" {
		body["transaction_id"] = req.ProviderTxnId
	}
	if req.Reason != "" {
		body["reason"] = req.Reason
	}
	b, _ := json.Marshal(body)
	reqh, err := http.NewRequestWithContext(ctx, "POST", urlStr, strings.NewReader(string(b)))
	if err != nil {
		return nil, err
	}
	reqh.Header.Set("Content-Type", "application/json")
	auth, err := w.sign("POST", "/v3/refund/domestic/refunds", string(b))
	if err != nil {
		return nil, err
	}
	reqh.Header.Set("Authorization", auth)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(reqh)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	rb, _ := ioutil.ReadAll(resp.Body)
	var r map[string]any
	_ = json.Unmarshal(rb, &r)
	rr := &payc.RefundResult{Success: true, OutRefundNo: req.OutRefundNo, RefundId: fmt.Sprintf("%v", r["refund_id"]), Raw: r}
	return rr, nil
}

func (w *Wechat) Capture(ctx context.Context, req *payc.CaptureRequest) (*payc.CaptureResult, error) {
	return nil, &payc.PaymentError{Code: "NotImplemented", Message: "wechat capture not implemented"}
}

func (w *Wechat) ParseNotify(ctx context.Context, headers map[string]string, body []byte) (*payc.NotifyEvent, error) {
	var msg map[string]any
	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, &payc.PaymentError{Code: "BadRequest", Message: err.Error()}
	}
	rsrc, _ := msg["resource"].(map[string]any)
	ciphertext, _ := rsrc["ciphertext"].(string)
	associatedData, _ := rsrc["associated_data"].(string)
	nonce, _ := rsrc["nonce"].(string)
	pt, err := wechatDecrypt(w.cfg.ApiV3Key, associatedData, nonce, ciphertext)
	if err != nil {
		return nil, &payc.PaymentError{Code: "DecryptFailed", Message: err.Error()}
	}
	var plain map[string]any
	_ = json.Unmarshal(pt, &plain)
	ev := &payc.NotifyEvent{EventType: fmt.Sprintf("%v", msg["event_type"]), IdempotencyKey: fmt.Sprintf("%v", plain["out_trade_no"]), Transaction: &payc.Transaction{Provider: "wechat", Raw: plain}}
	return ev, nil
}

func (w *Wechat) ProfitShare(ctx context.Context, req *payc.ProfitShareRequest) (*payc.ProfitShareResult, error) {
	urlStr := "https://api.mch.weixin.qq.com/v3/profitsharing/orders"
	receivers := make([]map[string]any, 0, len(req.Receivers))
	for _, r := range req.Receivers {
		receivers = append(receivers, map[string]any{"type": r.Type, "account": r.Account, "amount": r.Amount, "description": r.Desc})
	}
	body := map[string]any{"transaction_id": req.ProviderTxnId, "out_order_no": req.OutOrderNo, "receivers": receivers, "unfreeze_unsplit": req.UnfreezeUnsplit}
	b, _ := json.Marshal(body)
	reqh, err := http.NewRequestWithContext(ctx, "POST", urlStr, strings.NewReader(string(b)))
	if err != nil {
		return nil, err
	}
	reqh.Header.Set("Content-Type", "application/json")
	auth, err := w.sign("POST", "/v3/profitsharing/orders", string(b))
	if err != nil {
		return nil, err
	}
	reqh.Header.Set("Authorization", auth)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(reqh)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	rb, _ := ioutil.ReadAll(resp.Body)
	var r map[string]any
	_ = json.Unmarshal(rb, &r)
	return &payc.ProfitShareResult{Success: true, Raw: r}, nil
}

func (w *Wechat) CreateCombined(ctx context.Context, req *payc.CombinedCreateRequest) (*payc.ClientAction, error) {
	api := ""
	if req.Scene == payc.SceneNative {
		api = "/v3/combine-transactions/native"
	} else if req.Scene == payc.SceneJSAPI || req.Scene == payc.SceneMiniProgram {
		api = "/v3/combine-transactions/jsapi"
	} else if req.Scene == payc.SceneH5 {
		api = "/v3/combine-transactions/h5"
	} else {
		return nil, &payc.PaymentError{Code: "UnsupportedScene", Message: string(req.Scene)}
	}
	urlStr := "https://api.mch.weixin.qq.com" + api
	subs := make([]map[string]any, 0, len(req.SubOrders))
	for _, s := range req.SubOrders {
		item := map[string]any{
			"mchid":        s.MchID,
			"appid":        s.AppID,
			"out_trade_no": s.OutTradeNo,
			"amount":       map[string]any{"total": s.Amount},
			"description":  s.Description,
		}
		if s.PayerOpenID != "" {
			item["payer"] = map[string]any{"openid": s.PayerOpenID}
		}
		subs = append(subs, item)
	}
	body := map[string]any{"combine_out_trade_no": req.CombineOutTradeNo, "sub_orders": subs, "notify_url": req.NotifyURL}
	b, _ := json.Marshal(body)
	reqh, err := http.NewRequestWithContext(ctx, "POST", urlStr, strings.NewReader(string(b)))
	if err != nil {
		return nil, err
	}
	reqh.Header.Set("Content-Type", "application/json")
	auth, err := w.sign("POST", api, string(b))
	if err != nil {
		return nil, err
	}
	reqh.Header.Set("Authorization", auth)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(reqh)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	rb, _ := ioutil.ReadAll(resp.Body)
	var r map[string]any
	_ = json.Unmarshal(rb, &r)
	if req.Scene == payc.SceneNative {
		v := ""
		if x, ok := r["code_url"].(string); ok {
			v = x
		}
		return &payc.ClientAction{Kind: payc.ActionQRCode, Data: map[string]string{"code_url": v}}, nil
	}
	if req.Scene == payc.SceneJSAPI || req.Scene == payc.SceneMiniProgram {
		prepayId := ""
		if x, ok := r["prepay_id"].(string); ok {
			prepayId = x
		}
		ts := fmt.Sprintf("%d", time.Now().Unix())
		nonce := randStr(16)
		pkg := "prepay_id=" + prepayId
		paySign, err := w.paySign(ts, nonce, pkg)
		if err != nil {
			return nil, err
		}
		data := map[string]string{"appId": w.cfg.AppID, "timeStamp": ts, "nonceStr": nonce, "package": pkg, "signType": "RSA", "paySign": paySign}
		return &payc.ClientAction{Kind: payc.ActionMiniParams, Data: data}, nil
	}
	h5 := ""
	if x, ok := r["h5_url"].(string); ok {
		h5 = x
	}
	return &payc.ClientAction{Kind: payc.ActionRedirect, Data: map[string]string{"url": h5}}, nil
}

func (w *Wechat) sign(method string, canonicalURL string, body string) (string, error) {
	ts := fmt.Sprintf("%d", time.Now().Unix())
	nonce := randStr(16)
	message := strings.Join([]string{method, canonicalURL, ts, nonce, body}, "\n") + "\n"
	pk, err := ioutil.ReadFile(w.cfg.PrivateKeyPath)
	if err != nil {
		return "", err
	}
	block, _ := pem.Decode(pk)
	if block == nil {
		return "", fmt.Errorf("pem decode failed")
	}
	pri, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		pri, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return "", err
		}
	}
	rsaPri := pri.(*rsa.PrivateKey)
	h := sha256.New()
	h.Write([]byte(message))
	sig, err := rsaPri.Sign(rand.Reader, h.Sum(nil), crypto.SHA256)
	if err != nil {
		return "", err
	}
	s := base64.StdEncoding.EncodeToString(sig)
	auth := fmt.Sprintf("WECHATPAY2-SHA256-RSA2048 mchid=\"%s\",serial_no=\"%s\",nonce_str=\"%s\",timestamp=\"%s\",signature=\"%s\"", w.cfg.MchID, w.cfg.SerialNo, nonce, ts, s)
	return auth, nil
}

func (w *Wechat) paySign(ts, nonce, pkg string) (string, error) {
	msg := strings.Join([]string{w.cfg.AppID, ts, nonce, pkg}, "\n") + "\n"
	pk, err := ioutil.ReadFile(w.cfg.PrivateKeyPath)
	if err != nil {
		return "", err
	}
	block, _ := pem.Decode(pk)
	if block == nil {
		return "", fmt.Errorf("pem decode failed")
	}
	pri, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		pri, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return "", err
		}
	}
	rsaPri := pri.(*rsa.PrivateKey)
	h := sha256.New()
	h.Write([]byte(msg))
	sig, err := rsaPri.Sign(rand.Reader, h.Sum(nil), crypto.SHA256)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

func randStr(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[mathrand.Intn(len(letters))]
	}
	return string(b)
}

func wechatDecrypt(apiv3Key string, associatedData string, nonce string, ciphertext string) ([]byte, error) {
	key := []byte(apiv3Key)
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}
	return aesGCMDecrypt(key, []byte(associatedData), []byte(nonce), data)
}
