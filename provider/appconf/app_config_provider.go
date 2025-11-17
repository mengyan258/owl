package appconf

import (
	"bit-labs.cn/owl/contract/foundation"
	_ "embed"
)

type AppConfigServiceProvider struct {
	app foundation.Application
}

var _ foundation.ServiceProvider = (*AppConfigServiceProvider)(nil)

func (s AppConfigServiceProvider) Register() {

}

func (s AppConfigServiceProvider) Boot() {

}

//go:embed app.yaml
var appYaml string

func (s AppConfigServiceProvider) GenerateConf() map[string]string {
	return map[string]string{
		"app.yaml": appYaml,
	}
}
