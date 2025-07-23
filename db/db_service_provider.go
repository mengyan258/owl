package db

import (
	"bit-labs.cn/owl/conf"
	"bit-labs.cn/owl/contract/foundation"
	"bit-labs.cn/owl/contract/log"
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
		if err != nil {
			panic(err)
		}
		if opt.Driver == Sqlite {
			opt.Host = filepath.Join(i.app.ConfigPath(""), opt.Host)
			l.Debug("sqlite path:", opt.Host)
		}
		return InitDB(&opt)
	})
}

func (i *DBServiceProvider) Boot() {

}
