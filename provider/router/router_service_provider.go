package router

import (
	"bit-labs.cn/owl/contract/foundation"
	"bit-labs.cn/owl/provider/conf"
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
