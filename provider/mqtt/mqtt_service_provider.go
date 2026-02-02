package mqtt

import (
	_ "embed"

	"bit-labs.cn/owl/provider/conf"

	"bit-labs.cn/owl"
	"bit-labs.cn/owl/contract/foundation"
)

var _ foundation.ServiceProvider = (*MQTTServiceProvider)(nil)

type MQTTServiceProvider struct {
	app foundation.Application
}

func (m *MQTTServiceProvider) Description() string {
	return "MQTT 客户端连接与发布订阅"
}

func (m *MQTTServiceProvider) Register() {
	m.app.Register(func(c *conf.Configure) *MQTTClient {
		var opt Options
		err := c.GetConfig("mqtt", &opt)
		owl.PanicIf(err)

		return InitMQTT(&opt)
	})
}

func (m *MQTTServiceProvider) Boot() {

}

//go:embed mqtt.yaml
var mqttYaml string

func (m *MQTTServiceProvider) Conf() map[string]string {
	return map[string]string{
		"mqtt.yaml": mqttYaml,
	}
}
