package event

import (
	"bit-labs.cn/owl/contract/foundation"
	"bit-labs.cn/owl/contract/log"
	"github.com/asaskevich/EventBus"
)

var _ foundation.ServiceProvider = (*EventServiceProvider)(nil)

type EventServiceProvider struct {
	app foundation.Application
}

func (i *EventServiceProvider) Description() string {
	return "应用事件总线与发布订阅"
}

func (i *EventServiceProvider) Register() {
	i.app.Register(func() EventBus.Bus {
		return EventBus.New()
	})
}

func (i *EventServiceProvider) Boot() {
	err := i.app.Invoke(func(l log.Logger) {
		l.Info("EventServiceProvider Booted")
	})
	if err != nil {
		panic(err)
	}
}

func (i *EventServiceProvider) Conf() map[string]string {
	return nil
}
