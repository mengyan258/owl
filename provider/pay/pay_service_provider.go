package pay

import (
	_ "embed"
	"time"

	"bit-labs.cn/owl"
	"bit-labs.cn/owl/contract/foundation"
	"bit-labs.cn/owl/provider/conf"
	"bit-labs.cn/owl/provider/pay/impl"
	"gorm.io/gorm"
)

var _ foundation.ServiceProvider = (*PayServiceProvider)(nil)

type PayServiceProvider struct {
	app foundation.Application
}

func (p *PayServiceProvider) Description() string {
	return "支付驱动管理与默认支付通道"
}

func (p *PayServiceProvider) Register() {
	p.app.Register(func(c *conf.Configure) *PaymentManager {
		var opt Options
		err := c.GetConfig("pay", &opt)
		owl.PanicIf(err)

		m := NewPaymentManager()

		if opt.Alipay.AppID != "" {
			if d, e := impl.NewAlipay(&opt.Alipay); e == nil {
				m.AddDriver("alipay", d)
			}
		}
		if opt.Wechat.MchID != "" {
			if d, e := impl.NewWechat(&opt.Wechat); e == nil {
				m.AddDriver("wechat", d)
			}
		}
		if opt.Card.Gateway != "" {
			if d, e := impl.NewCard(&opt.Card); e == nil {
				m.AddDriver("card", d)
			}
		}

		def := opt.Default
		if def == "" {
			if _, e := m.GetDriver("wechat"); e == nil {
				def = "wechat"
			} else if _, e := m.GetDriver("alipay"); e == nil {
				def = "alipay"
			} else if _, e := m.GetDriver("card"); e == nil {
				def = "card"
			}
		}
		if def != "" {
			err = m.SetDefaultDriver(def)
			owl.PanicIf(err)
		}

		return m
	})
}

func (p *PayServiceProvider) Boot() {
	_ = p.app.Invoke(func(c *conf.Configure, db *gorm.DB) {
		var opt Options
		if err := c.GetConfig("pay", &opt); err == nil {
			ttl := time.Duration(opt.DedupTTL) * time.Second
			if ttl <= 0 {
				ttl = 24 * time.Hour
			}
			InitDBDedupStore(db, ttl)
		}
	})
}

//go:embed pay.yaml
var payYaml string

func (p *PayServiceProvider) Conf() map[string]string {
	return map[string]string{
		"pay.yaml": payYaml,
	}
}
