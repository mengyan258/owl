package pay

import (
	"context"
	"time"
)

type PayScene string

const (
	SceneAPP         PayScene = "APP"
	SceneH5          PayScene = "H5"
	ScenePCQR        PayScene = "PC_QR"
	SceneJSAPI       PayScene = "JSAPI"
	SceneMiniProgram PayScene = "MINI_PROGRAM"
	SceneNative      PayScene = "NATIVE"
	SceneMicroPay    PayScene = "MICROPAY"
	ScenePreAuth     PayScene = "PREAUTH"
)

type PayChannel string

const (
	ChannelAlipay PayChannel = "ALIPAY"
	ChannelWechat PayChannel = "WECHAT"
	ChannelCard   PayChannel = "CARD"
)

type ClientActionKind string

const (
	ActionRedirect   ClientActionKind = "redirect"
	ActionQRCode     ClientActionKind = "qrcode"
	ActionInvokeSDK  ClientActionKind = "invokeSDK"
	ActionDeepLink   ClientActionKind = "deeplink"
	ActionMiniParams ClientActionKind = "miniProgramParams"
)

type ClientAction struct {
	Kind ClientActionKind  `json:"kind"`
	Data map[string]string `json:"data"`
}

type PayerInfo struct {
	OpenID    string `json:"open_id"`
	UserID    string `json:"user_id"`
	CardToken string `json:"card_token"`
}

type SplitInfo struct {
	Receiver string `json:"receiver"`
	Amount   int64  `json:"amount"`
	Desc     string `json:"desc"`
}

type PayIntent struct {
	OutTradeNo  string            `json:"out_trade_no"`
	Amount      int64             `json:"amount"`
	Currency    string            `json:"currency"`
	Subject     string            `json:"subject"`
	Description string            `json:"description"`
	Scene       PayScene          `json:"scene"`
	Channel     PayChannel        `json:"channel"`
	NotifyURL   string            `json:"notify_url"`
	ReturnURL   string            `json:"return_url"`
	ExpireAt    time.Time         `json:"expire_at"`
	Payer       PayerInfo         `json:"payer"`
	Attach      map[string]string `json:"attach"`
	Extra       map[string]string `json:"extra"`
	Split       []SplitInfo       `json:"split"`
}

type TransactionStatus string

const (
	TxCreated           TransactionStatus = "CREATED"
	TxPending           TransactionStatus = "PENDING"
	TxPaid              TransactionStatus = "PAID"
	TxAuthorized        TransactionStatus = "AUTHORIZED"
	TxCaptured          TransactionStatus = "CAPTURED"
	TxClosed            TransactionStatus = "CLOSED"
	TxRefunded          TransactionStatus = "REFUNDED"
	TxPartiallyRefunded TransactionStatus = "PARTIALLY_REFUNDED"
	TxFailed            TransactionStatus = "FAILED"
	TxExpired           TransactionStatus = "EXPIRED"
)

type Transaction struct {
	Status        TransactionStatus `json:"status"`
	Provider      string            `json:"provider"`
	Channel       PayChannel        `json:"channel"`
	Scene         PayScene          `json:"scene"`
	OutTradeNo    string            `json:"out_trade_no"`
	ProviderTxnId string            `json:"provider_txn_id"`
	Amount        int64             `json:"amount"`
	Currency      string            `json:"currency"`
	Payer         PayerInfo         `json:"payer"`
	Raw           map[string]any    `json:"raw"`
}

type QueryRequest struct {
	OutTradeNo    string `json:"out_trade_no"`
	ProviderTxnId string `json:"provider_txn_id"`
}

type CloseRequest struct {
	OutTradeNo string `json:"out_trade_no"`
}

type RefundRequest struct {
	OutRefundNo   string            `json:"out_refund_no"`
	OutTradeNo    string            `json:"out_trade_no"`
	ProviderTxnId string            `json:"provider_txn_id"`
	Amount        int64             `json:"amount"`
	Reason        string            `json:"reason"`
	Attach        map[string]string `json:"attach"`
}

type RefundResult struct {
	Success     bool           `json:"success"`
	OutRefundNo string         `json:"out_refund_no"`
	RefundId    string         `json:"refund_id"`
	Raw         map[string]any `json:"raw"`
}

type CaptureRequest struct {
	OutTradeNo    string `json:"out_trade_no"`
	ProviderTxnId string `json:"provider_txn_id"`
	Amount        int64  `json:"amount"`
}

type CaptureResult struct {
	Success bool           `json:"success"`
	Raw     map[string]any `json:"raw"`
}

type NotifyEvent struct {
	EventType      string        `json:"event_type"`
	Transaction    *Transaction  `json:"transaction"`
	Refund         *RefundResult `json:"refund"`
	IdempotencyKey string        `json:"idempotency_key"`
}

type PaymentError struct {
	Code        string
	Message     string
	Retryable   bool
	ProviderRaw map[string]any
}

func (e *PaymentError) Error() string {
	return e.Message
}

type PayDriver interface {
	Create(ctx context.Context, intent *PayIntent) (*ClientAction, *Transaction, error)
	Query(ctx context.Context, req *QueryRequest) (*Transaction, error)
	Close(ctx context.Context, req *CloseRequest) error
	Refund(ctx context.Context, req *RefundRequest) (*RefundResult, error)
	Capture(ctx context.Context, req *CaptureRequest) (*CaptureResult, error)
	ParseNotify(ctx context.Context, headers map[string]string, body []byte) (*NotifyEvent, error)
	ProfitShare(ctx context.Context, req *ProfitShareRequest) (*ProfitShareResult, error)
	CreateCombined(ctx context.Context, req *CombinedCreateRequest) (*ClientAction, error)
}

type ProfitShareReceiver struct {
	Type    string `json:"type"`
	Account string `json:"account"`
	Amount  int64  `json:"amount"`
	Desc    string `json:"desc"`
}

type ProfitShareRequest struct {
	OutTradeNo      string                `json:"out_trade_no"`
	ProviderTxnId   string                `json:"provider_txn_id"`
	OutOrderNo      string                `json:"out_order_no"`
	Receivers       []ProfitShareReceiver `json:"receivers"`
	UnfreezeUnsplit bool                  `json:"unfreeze_unsplit"`
}

type ProfitShareResult struct {
	Success bool           `json:"success"`
	Raw     map[string]any `json:"raw"`
}

type CombinedSubOrder struct {
	MchID       string `json:"mchid"`
	AppID       string `json:"appid"`
	OutTradeNo  string `json:"out_trade_no"`
	Amount      int64  `json:"amount"`
	Description string `json:"description"`
	PayerOpenID string `json:"payer_openid"`
}

type CombinedCreateRequest struct {
	CombineOutTradeNo string             `json:"combine_out_trade_no"`
	Scene             PayScene           `json:"scene"`
	SubOrders         []CombinedSubOrder `json:"sub_orders"`
	NotifyURL         string             `json:"notify_url"`
}
