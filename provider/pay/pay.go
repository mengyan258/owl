package pay

import (
	"context"
	"fmt"

	payc "bit-labs.cn/owl/contract/pay"
)

type Options struct {
	Default  string            `json:"default"`
	Alipay   payc.AlipayConfig `json:"alipay"`
	Wechat   payc.WechatConfig `json:"wechat"`
	Card     payc.CardConfig   `json:"card"`
	DedupTTL int               `json:"dedup_ttl"`
}

type PaymentManager struct {
	drivers map[string]payc.PayDriver
	def     string
	opt     *Options
}

func NewPaymentManager() *PaymentManager {
	return &PaymentManager{drivers: make(map[string]payc.PayDriver)}
}

func (m *PaymentManager) AddDriver(name string, d payc.PayDriver) {
	m.drivers[name] = d
}

func (m *PaymentManager) SetDefaultDriver(name string) error {
	if _, ok := m.drivers[name]; !ok {
		return fmt.Errorf("pay driver '%s' not found", name)
	}
	m.def = name
	return nil
}

func (m *PaymentManager) GetDriver(name string) (payc.PayDriver, error) {
	if name == "" {
		name = m.def
	}
	d, ok := m.drivers[name]
	if !ok {
		return nil, fmt.Errorf("pay driver %s not found", name)
	}
	return d, nil
}

func (m *PaymentManager) Default() (payc.PayDriver, error) {
	return m.GetDriver(m.def)
}

func (m *PaymentManager) Use(name string) (payc.PayDriver, error) {
	return m.GetDriver(name)
}

func (m *PaymentManager) Create(ctx context.Context, intent *payc.PayIntent) (*payc.ClientAction, *payc.Transaction, error) {
	if intent == nil || intent.Amount <= 0 {
		return nil, nil, fmt.Errorf("invalid intent")
	}
	if intent.Currency == "" {
		intent.Currency = "CNY"
	}
	d, err := m.Default()
	if err != nil {
		return nil, nil, err
	}
	return d.Create(ctx, intent)
}

func (m *PaymentManager) Query(ctx context.Context, req *payc.QueryRequest) (*payc.Transaction, error) {
	d, err := m.Default()
	if err != nil {
		return nil, err
	}
	return d.Query(ctx, req)
}

func (m *PaymentManager) Close(ctx context.Context, req *payc.CloseRequest) error {
	d, err := m.Default()
	if err != nil {
		return err
	}
	return d.Close(ctx, req)
}

func (m *PaymentManager) Refund(ctx context.Context, req *payc.RefundRequest) (*payc.RefundResult, error) {
	d, err := m.Default()
	if err != nil {
		return nil, err
	}
	return d.Refund(ctx, req)
}

func (m *PaymentManager) Capture(ctx context.Context, req *payc.CaptureRequest) (*payc.CaptureResult, error) {
	d, err := m.Default()
	if err != nil {
		return nil, err
	}
	return d.Capture(ctx, req)
}

func (m *PaymentManager) ProfitShare(ctx context.Context, req *payc.ProfitShareRequest) (*payc.ProfitShareResult, error) {
	d, err := m.Default()
	if err != nil {
		return nil, err
	}
	return d.ProfitShare(ctx, req)
}

func (m *PaymentManager) CreateCombined(ctx context.Context, req *payc.CombinedCreateRequest) (*payc.ClientAction, error) {
	d, err := m.Default()
	if err != nil {
		return nil, err
	}
	return d.CreateCombined(ctx, req)
}
