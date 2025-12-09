package db

import (
	_ "embed"
	"path/filepath"

	"bit-labs.cn/owl"
	"bit-labs.cn/owl/contract/foundation"
	"bit-labs.cn/owl/contract/log"
	"bit-labs.cn/owl/provider/conf"
	"gorm.io/gorm"
)

var _ foundation.ServiceProvider = (*DBServiceProvider)(nil)

type DBServiceProvider struct {
	app foundation.Application
}

func (i *DBServiceProvider) Register() {

	i.app.Register(func(c *conf.Configure, l log.Logger) *gorm.DB {

		var opt Options
		err := c.GetConfig("database", &opt)
		owl.PanicIf(err)

		if opt.Driver == Sqlite {
			opt.Host = filepath.Join(i.app.GetConfigPath(), opt.Host)
			l.Debug("use sqlite, path:", opt.Host)
		}
		return InitDB(&opt)
	})
}

func (i *DBServiceProvider) Boot() {

}

//go:embed database.yaml
var databaseYaml string

func (i *DBServiceProvider) GenerateConf() map[string]string {
	return map[string]string{
		"database.yaml": databaseYaml,
	}
}
