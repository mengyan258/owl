package db

import (
	"bit-labs.cn/owl"
	"bit-labs.cn/owl/contract/foundation"
	"bit-labs.cn/owl/contract/log"
	"bit-labs.cn/owl/provider/conf"
	"gorm.io/gorm"
	"path/filepath"
)

var _ foundation.ServiceProvider = (*DBServiceProvider)(nil)

type DBServiceProvider struct {
	app foundation.Application
}

func (i *DBServiceProvider) Register() {

	i.app.Register(func(c *conf.Configure, l log.Logger) *gorm.DB {

		var opt Options
		err := c.GetConfig("database.db", &opt)
		owl.PanicIf(err)

		if opt.Driver == Sqlite {
			opt.Host = filepath.Join(i.app.ConfigPath(""), opt.Host)
			l.Debug("use sqlite, path:", opt.Host)
		}
		return InitDB(&opt)
	})
}

func (i *DBServiceProvider) Boot() {

}
