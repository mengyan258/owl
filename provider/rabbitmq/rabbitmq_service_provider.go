package rabbitmq

import (
	_ "embed"

	"bit-labs.cn/owl"
	"bit-labs.cn/owl/contract/foundation"
	"bit-labs.cn/owl/provider/conf"
)

var _ foundation.ServiceProvider = (*RabbitMQServiceProvider)(nil)

type RabbitMQServiceProvider struct {
	app foundation.Application
}

func NewRabbitMQServiceProvider(app foundation.Application) *RabbitMQServiceProvider {
	return &RabbitMQServiceProvider{
		app: app,
	}
}

func (r *RabbitMQServiceProvider) Register() {
	r.app.Register(func(c *conf.Configure) *RabbitMQClient {
		var opt Options
		err := c.GetConfig("rabbitmq", &opt)
		owl.PanicIf(err)

		return InitRabbitMQ(&opt)
	})
}

func (r *RabbitMQServiceProvider) Boot() {
}

//go:embed rabbitmq.yaml
var rabbitmqYaml string

func (r *RabbitMQServiceProvider) GenerateConf() map[string]string {
	return map[string]string{
		"rabbitmq.yaml": rabbitmqYaml,
	}
}
