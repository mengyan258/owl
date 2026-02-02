package conf

import (
	"bit-labs.cn/owl/contract/foundation"
	"github.com/asaskevich/EventBus"
)

var _ foundation.ServiceProvider = (*ConfServiceProvider)(nil)

type ConfServiceProvider struct {
	app foundation.Application
}

func (i *ConfServiceProvider) Description() string {
	return "核心服务提供者，自动加载 conf 下的所有配置文件，支持 json，yaml，tomal"
}

func (i *ConfServiceProvider) Register() {

	i.app.Register(func(bus EventBus.Bus) *Configure {
		return NewConfigure(i.app, bus)
	})
}

func (i *ConfServiceProvider) Boot() {

}

func (i *ConfServiceProvider) Conf() map[string]string {
	return nil
}
