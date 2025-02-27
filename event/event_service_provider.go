package event

import (
	"bit-labs.cn/owl/contract/foundation"
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
}
