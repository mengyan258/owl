package middleware

import (
	"net/http"
	"runtime/debug"

	logContract "bit-labs.cn/owl/contract/log"
	"github.com/gin-gonic/gin"
)

func Recovery(logger logContract.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				requestID, _ := c.Get("request_id")
				logger.Emergency(
					"HTTP panic",
					" requestId:", requestID,
					" method:", c.Request.Method,
					" path:", c.Request.URL.Path,
					" panic info:", recovered,
					" stack", string(debug.Stack()),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"msg":     "服务器内部错误",
				})
			}
		}()
		c.Next()
	}
}
