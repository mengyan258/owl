package db

import (
	"bit-labs.cn/owl/conf"
	"bit-labs.cn/owl/contract/foundation"
	"gorm.io/gorm"
)

var _ foundation.ServiceProvider = (*DBServiceProvider)(nil)

type DBServiceProvider struct {
	app foundation.Application
}

func (i *DBServiceProvider) Register() {

	i.app.Register(func(c *conf.Configure) *gorm.DB {
		var opt Options
		err := c.GetConfig("database.db", &opt)
		if err != nil {
			panic(err)
		}
		return InitDB(&opt)
	})
}

func (i *DBServiceProvider) Boot() {

}
