package mqtt

import (
	"bit-labs.cn/owl/provider/conf"
	_ "embed"

	"bit-labs.cn/owl"
	"bit-labs.cn/owl/contract/foundation"
)

var _ foundation.ServiceProvider = (*MQTTServiceProvider)(nil)

type MQTTServiceProvider struct {
	app foundation.Application
}

func NewMQTTServiceProvider(app foundation.Application) *MQTTServiceProvider {
	return &MQTTServiceProvider{
		app: app,
	}
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

func (m *MQTTServiceProvider) GenerateConf() map[string]string {
	return map[string]string{
		"mqtt.yaml": mqttYaml,
	}
}
