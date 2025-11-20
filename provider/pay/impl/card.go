package impl

import (
	payc "bit-labs.cn/owl/contract/pay"
	"context"
)

type Card struct {
	cfg *payc.CardConfig
}

func NewCard(cfg *payc.CardConfig) (*Card, error) {
	return &Card{cfg: cfg}, nil
}

func (c *Card) Create(ctx context.Context, intent *payc.PayIntent) (*payc.ClientAction, *payc.Transaction, error) {
	return nil, nil, &payc.PaymentError{Code: "NotImplemented", Message: "card create not implemented"}
}

func (c *Card) Query(ctx context.Context, req *payc.QueryRequest) (*payc.Transaction, error) {
	return nil, &payc.PaymentError{Code: "NotImplemented", Message: "card query not implemented"}
}

func (c *Card) Close(ctx context.Context, req *payc.CloseRequest) error {
	return &payc.PaymentError{Code: "NotImplemented", Message: "card close not implemented"}
}

func (c *Card) Refund(ctx context.Context, req *payc.RefundRequest) (*payc.RefundResult, error) {
	return nil, &payc.PaymentError{Code: "NotImplemented", Message: "card refund not implemented"}
}

func (c *Card) Capture(ctx context.Context, req *payc.CaptureRequest) (*payc.CaptureResult, error) {
	return nil, &payc.PaymentError{Code: "NotImplemented", Message: "card capture not implemented"}
}

func (c *Card) ParseNotify(ctx context.Context, headers map[string]string, body []byte) (*payc.NotifyEvent, error) {
	return nil, &payc.PaymentError{Code: "NotImplemented", Message: "card notify not implemented"}
}

func (c *Card) ProfitShare(ctx context.Context, req *payc.ProfitShareRequest) (*payc.ProfitShareResult, error) {
	return nil, &payc.PaymentError{Code: "NotImplemented", Message: "card profitsharing not implemented"}
}

func (c *Card) CreateCombined(ctx context.Context, req *payc.CombinedCreateRequest) (*payc.ClientAction, error) {
	return nil, &payc.PaymentError{Code: "NotImplemented", Message: "card combined not implemented"}
}
