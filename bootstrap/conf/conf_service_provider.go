package conf

import (
	"bit-labs.cn/owl/contract/foundation"
	"github.com/asaskevich/EventBus"
)

var _ foundation.ServiceProvider = (*ConfServiceProvider)(nil)

type ConfServiceProvider struct {
	app foundation.Application
}

func (i *ConfServiceProvider) Register() {

	i.app.Register(func(bus EventBus.Bus) *Configure {
		return NewConfigure(i.app, bus)
	})
}

func (i *ConfServiceProvider) Boot() {

}

func (i *ConfServiceProvider) GenerateConf() map[string]string {
	return nil
}
