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

func (i *EventServiceProvider) GenerateConf() map[string]string {
	return nil
}
