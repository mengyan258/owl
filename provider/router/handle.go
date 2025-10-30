package router

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Handler 控制器需要实现此接口，用于获取模块名称，用于设置 api 接口信息
type Handler interface {
	ModuleName() (en string, zh string)
}

type CrudHandler interface {
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
	Retrieve(ctx *gin.Context)
	Detail(ctx *gin.Context)
}

func Success(ctx *gin.Context, data any) {
	ctx.JSON(http.StatusOK, gin.H{"success": true, "msg": "操作成功", "data": data})
}

func Fail(ctx *gin.Context, err string) {
	ctx.JSON(http.StatusOK, gin.H{"success": false, "msg": err, "data": ""})
}

func Auto(ctx *gin.Context, data any, err error) {
	if err != nil {
		Fail(ctx, err.Error())
	} else {
		Success(ctx, data)
	}
}
