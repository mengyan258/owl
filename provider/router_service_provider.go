package provider

import (
	"bit-labs.cn/owl/conf"
	"bit-labs.cn/owl/contract/foundation"
	"github.com/gin-gonic/gin"
)

var _ foundation.ServiceProvider = (*RouterServiceProvider)(nil)

type RouterServiceProvider struct {
	app foundation.Application
}

func (i *RouterServiceProvider) Register() {
	i.app.Register(func(c *conf.Configure) *gin.Engine {
		return gin.New()
	})
}

func (i *RouterServiceProvider) Boot() {

}
