package impl

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	payc "bit-labs.cn/owl/contract/pay"
)

type Alipay struct {
	cfg *payc.AlipayConfig
}

func NewAlipay(cfg *payc.AlipayConfig) (*Alipay, error) {
	return &Alipay{cfg: cfg}, nil
}

func (a *Alipay) Create(ctx context.Context, intent *payc.PayIntent) (*payc.ClientAction, *payc.Transaction, error) {
	if intent == nil {
		return nil, nil, &payc.PaymentError{Code: "Invalid", Message: "intent nil"}
	}
	m := ""
	extra := map[string]string{}
	switch intent.Scene {
	case payc.SceneH5:
		m = "alipay.trade.wap.pay"
		extra["product_code"] = "QUICK_WAP_WAY"
	case payc.ScenePCQR:
		m = "alipay.trade.precreate"
	case payc.SceneAPP:
		m = "alipay.trade.app.pay"
		extra["product_code"] = "QUICK_MSECURITY_PAY"
	default:
		return nil, nil, &payc.PaymentError{Code: "UnsupportedScene", Message: string(intent.Scene)}
	}
	biz := map[string]any{
		"subject":      intent.Subject,
		"out_trade_no": intent.OutTradeNo,
		"total_amount": fmt.Sprintf("%.2f", float64(intent.Amount)/100.0),
	}
	for k, v := range intent.Extra {
		biz[k] = v
	}
	b, _ := json.Marshal(biz)
	if intent.Scene == payc.ScenePCQR {
		r, err := a.doRequest(m, string(b), extra)
		if err != nil {
			return nil, nil, err
		}
		resp := a.extractResponse(m, r)
		qr := ""
		if v, ok := resp["qr_code"].(string); ok {
			qr = v
		}
		ca := &payc.ClientAction{Kind: payc.ActionQRCode, Data: map[string]string{"qr_code": qr}}
		tx := &payc.Transaction{Status: payc.TxCreated, Provider: "alipay", Channel: payc.ChannelAlipay, Scene: intent.Scene, OutTradeNo: intent.OutTradeNo, Amount: intent.Amount, Currency: intent.Currency, Payer: intent.Payer, Raw: resp}
		return ca, tx, nil
	}
	params := map[string]string{
		"app_id":      a.cfg.AppID,
		"method":      m,
		"format":      "JSON",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"biz_content": string(b),
	}
	if a.cfg.NotifyURL != "" {
		params["notify_url"] = a.cfg.NotifyURL
	}
	if a.cfg.ReturnURL != "" && m != "alipay.trade.app.pay" {
		params["return_url"] = a.cfg.ReturnURL
	}
	for k, v := range extra {
		params[k] = v
	}
	sign, err := a.sign(params)
	if err != nil {
		return nil, nil, err
	}
	params["sign"] = sign
	values := url.Values{}
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		values.Add(k, params[k])
	}
	gw := "https://openapi.alipay.com/gateway.do"
	if a.cfg.Sandbox {
		gw = "https://openapi.alipaydev.com/gateway.do"
	}
	if m == "alipay.trade.app.pay" {
		orderStr := values.Encode()
		ca := &payc.ClientAction{Kind: payc.ActionInvokeSDK, Data: map[string]string{"orderString": orderStr}}
		tx := &payc.Transaction{Status: payc.TxCreated, Provider: "alipay", Channel: payc.ChannelAlipay, Scene: intent.Scene, OutTradeNo: intent.OutTradeNo, Amount: intent.Amount, Currency: intent.Currency, Payer: intent.Payer}
		return ca, tx, nil
	}
	urlStr := gw + "?" + values.Encode()
	ca := &payc.ClientAction{Kind: payc.ActionRedirect, Data: map[string]string{"url": urlStr}}
	tx := &payc.Transaction{Status: payc.TxCreated, Provider: "alipay", Channel: payc.ChannelAlipay, Scene: intent.Scene, OutTradeNo: intent.OutTradeNo, Amount: intent.Amount, Currency: intent.Currency, Payer: intent.Payer}
	return ca, tx, nil
}

func (a *Alipay) Query(ctx context.Context, req *payc.QueryRequest) (*payc.Transaction, error) {
	biz := map[string]any{}
	if req.OutTradeNo != "" {
		biz["out_trade_no"] = req.OutTradeNo
	}
	if req.ProviderTxnId != "" {
		biz["trade_no"] = req.ProviderTxnId
	}
	b, _ := json.Marshal(biz)
	r, err := a.doRequest("alipay.trade.query", string(b), nil)
	if err != nil {
		return nil, err
	}
	resp := a.extractResponse("alipay.trade.query", r)
	status := payc.TxPending
	ts := fmt.Sprintf("%v", resp["trade_status"])
	switch ts {
	case "TRADE_SUCCESS":
		status = payc.TxPaid
	case "TRADE_FINISHED":
		status = payc.TxClosed
	case "TRADE_CLOSED":
		status = payc.TxClosed
	default:
		status = payc.TxPending
	}
	tx := &payc.Transaction{Status: status, Provider: "alipay", OutTradeNo: req.OutTradeNo, ProviderTxnId: fmt.Sprintf("%v", resp["trade_no"]), Raw: resp}
	return tx, nil
}

func (a *Alipay) Close(ctx context.Context, req *payc.CloseRequest) error {
	biz := map[string]any{"out_trade_no": req.OutTradeNo}
	b, _ := json.Marshal(biz)
	r, err := a.doRequest("alipay.trade.close", string(b), nil)
	if err != nil {
		return err
	}
	resp := a.extractResponse("alipay.trade.close", r)
	code := fmt.Sprintf("%v", resp["code"])
	if code != "10000" && code != "0" {
		return &payc.PaymentError{Code: code, Message: fmt.Sprintf("close failed: %v", resp["msg"])}
	}
	return nil
}

func (a *Alipay) Refund(ctx context.Context, req *payc.RefundRequest) (*payc.RefundResult, error) {
	biz := map[string]any{
		"refund_amount":  fmt.Sprintf("%.2f", float64(req.Amount)/100.0),
		"out_request_no": req.OutRefundNo,
	}
	if req.OutTradeNo != "" {
		biz["out_trade_no"] = req.OutTradeNo
	}
	if req.ProviderTxnId != "" {
		biz["trade_no"] = req.ProviderTxnId
	}
	if req.Reason != "" {
		biz["refund_reason"] = req.Reason
	}
	b, _ := json.Marshal(biz)
	r, err := a.doRequest("alipay.trade.refund", string(b), nil)
	if err != nil {
		return nil, err
	}
	resp := a.extractResponse("alipay.trade.refund", r)
	code := fmt.Sprintf("%v", resp["code"])
	if code != "10000" && code != "0" {
		return nil, &payc.PaymentError{Code: code, Message: fmt.Sprintf("refund failed: %v", resp["msg"])}
	}
	rr := &payc.RefundResult{Success: true, OutRefundNo: req.OutRefundNo, RefundId: req.OutRefundNo, Raw: resp}
	return rr, nil
}

func (a *Alipay) Capture(ctx context.Context, req *payc.CaptureRequest) (*payc.CaptureResult, error) {
	return nil, &payc.PaymentError{Code: "NotImplemented", Message: "alipay capture not implemented"}
}

func (a *Alipay) ParseNotify(ctx context.Context, headers map[string]string, body []byte) (*payc.NotifyEvent, error) {
	vals, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, &payc.PaymentError{Code: "BadRequest", Message: err.Error()}
	}
	sign := vals.Get("sign")
	vals.Del("sign")
	vals.Del("sign_type")
	list := make([]string, 0)
	for k := range vals {
		list = append(list, k)
	}
	sort.Strings(list)
	var sb strings.Builder
	for i, k := range list {
		if vals.Get(k) == "" {
			continue
		}
		if i > 0 && sb.Len() > 0 {
			sb.WriteString("&")
		}
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(vals.Get(k))
	}
	ok, verr := a.verify(sb.String(), sign)
	if verr != nil || !ok {
		return nil, &payc.PaymentError{Code: "VerifyFailed", Message: "verify failed"}
	}
	ev := &payc.NotifyEvent{EventType: "payment.succeeded", IdempotencyKey: vals.Get("out_trade_no")}
	return ev, nil
}

func (a *Alipay) ProfitShare(ctx context.Context, req *payc.ProfitShareRequest) (*payc.ProfitShareResult, error) {
	royalties := make([]map[string]any, 0, len(req.Receivers))
	for _, r := range req.Receivers {
		royalties = append(royalties, map[string]any{"trans_in": r.Account, "amount": fmt.Sprintf("%.2f", float64(r.Amount)/100.0), "desc": r.Desc})
	}
	biz := map[string]any{"out_request_no": req.OutOrderNo, "trade_no": req.ProviderTxnId, "royalty_parameters": royalties}
	b, _ := json.Marshal(biz)
	r, err := a.doRequest("alipay.trade.order.settle", string(b), nil)
	if err != nil {
		return nil, err
	}
	resp := a.extractResponse("alipay.trade.order.settle", r)
	return &payc.ProfitShareResult{Success: true, Raw: resp}, nil
}

func (a *Alipay) CreateCombined(ctx context.Context, req *payc.CombinedCreateRequest) (*payc.ClientAction, error) {
	return nil, &payc.PaymentError{Code: "NotImplemented", Message: "alipay combined not implemented"}
}

func (a *Alipay) doRequest(method string, bizContent string, extra map[string]string) (map[string]any, error) {
	params := map[string]string{
		"app_id":      a.cfg.AppID,
		"method":      method,
		"format":      "JSON",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"biz_content": bizContent,
	}
	if a.cfg.NotifyURL != "" {
		params["notify_url"] = a.cfg.NotifyURL
	}
	for k, v := range extra {
		params[k] = v
	}
	sign, err := a.sign(params)
	if err != nil {
		return nil, err
	}
	params["sign"] = sign
	values := url.Values{}
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		values.Add(k, params[k])
	}
	gw := "https://openapi.alipay.com/gateway.do"
	if a.cfg.Sandbox {
		gw = "https://openapi.alipaydev.com/gateway.do"
	}
	resp, err := httpPostForm(gw, values.Encode())
	if err != nil {
		return nil, err
	}
	var obj map[string]any
	_ = json.Unmarshal(resp, &obj)
	return obj, nil
}

func (a *Alipay) extractResponse(method string, obj map[string]any) map[string]any {
	key := strings.ReplaceAll(method, ".", "_") + "_response"
	if v, ok := obj[key].(map[string]any); ok {
		return v
	}
	return obj
}

func httpPostForm(urlStr string, body string) ([]byte, error) {
	req, err := http.NewRequest("POST", urlStr, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (a *Alipay) sign(params map[string]string) (string, error) {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for i, k := range keys {
		v := params[k]
		if v == "" || k == "sign" {
			continue
		}
		if i > 0 && sb.Len() > 0 {
			sb.WriteString("&")
		}
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(v)
	}
	data := sb.String()
	pk, err := ioutil.ReadFile(a.cfg.PrivateKeyPath)
	if err != nil {
		return "", err
	}
	block, _ := pem.Decode(pk)
	if block == nil {
		return "", fmt.Errorf("pem decode failed")
	}
	pri, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		k2, e2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if e2 != nil {
			return "", err
		}
		pri = k2.(*rsa.PrivateKey)
	}
	h := sha256.New()
	h.Write([]byte(data))
	sig, err := rsa.SignPKCS1v15(nil, pri, crypto.SHA256, h.Sum(nil))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

func (a *Alipay) verify(data string, sign string) (bool, error) {
	pubBytes, err := ioutil.ReadFile(a.cfg.AlipayPublicKeyPath)
	if err != nil {
		return false, err
	}
	block, _ := pem.Decode(pubBytes)
	if block == nil {
		return false, fmt.Errorf("pem decode failed")
	}
	pubAny, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return false, err
	}
	pub := pubAny.(*rsa.PublicKey)
	b, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return false, err
	}
	h := sha256.New()
	h.Write([]byte(data))
	err = rsa.VerifyPKCS1v15(pub, crypto.SHA256, h.Sum(nil), b)
	if err != nil {
		return false, nil
	}
	return true, nil
}
