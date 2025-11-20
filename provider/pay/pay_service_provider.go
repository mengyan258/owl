package pay

import (
	"bit-labs.cn/owl"
	"bit-labs.cn/owl/contract/foundation"
	"bit-labs.cn/owl/provider/conf"
	"bit-labs.cn/owl/provider/pay/impl"
	_ "embed"
	"gorm.io/gorm"
	"time"
)

var _ foundation.ServiceProvider = (*PayServiceProvider)(nil)

type PayServiceProvider struct {
	app foundation.Application
}

func NewPayServiceProvider(app foundation.Application) *PayServiceProvider {
	return &PayServiceProvider{app: app}
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

//go:embed pay.yaml
var payYaml string

func (p *PayServiceProvider) GenerateConf() map[string]string {
	return map[string]string{
		"pay.yaml": payYaml,
	}
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
