package appconf

import (
	_ "embed"

	"bit-labs.cn/owl/contract/foundation"
)

type AppConfigServiceProvider struct {
	app foundation.Application
}

func (s AppConfigServiceProvider) Description() string {
	return "应用配置加载与管理"
}

var _ foundation.ServiceProvider = (*AppConfigServiceProvider)(nil)

func (s AppConfigServiceProvider) Register() {

}

func (s AppConfigServiceProvider) Boot() {

}

//go:embed app.yaml
var appYaml string

func (s AppConfigServiceProvider) Conf() map[string]string {
	return map[string]string{
		"app.yaml": appYaml,
	}
}
