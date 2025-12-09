package rabbitmq

import (
	_ "embed"

	"bit-labs.cn/owl/contract/log"
	"bit-labs.cn/owl/provider/conf"

	"bit-labs.cn/owl"
	"bit-labs.cn/owl/contract/foundation"
)

var _ foundation.ServiceProvider = (*RabbitMQServiceProvider)(nil)

type RabbitMQServiceProvider struct {
	app foundation.Application
}

func (r *RabbitMQServiceProvider) Register() {
	r.app.Register(func(c *conf.Configure, l log.Logger) *RabbitMQClient {
		var opt Options
		err := c.GetConfig("rabbitmq", &opt)
		owl.PanicIf(err)

		return NewRabbitMQ(&opt, l)
	})
}

func (r *RabbitMQServiceProvider) Boot() {
	r.app.Invoke(func(client *RabbitMQClient) {
		client.Connect()
	})
}

//go:embed rabbitmq.yaml
var rabbitmqYaml string

func (r *RabbitMQServiceProvider) GenerateConf() map[string]string {
	return map[string]string{
		"rabbitmq.yaml": rabbitmqYaml,
	}
}
