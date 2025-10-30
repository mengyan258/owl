package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

// RequestID 请求ID中间件
// 为每个请求生成唯一的请求ID，用于日志追踪和调试
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从请求头获取请求ID
		requestID := c.GetHeader("X-Request-ID")

		// 如果没有请求ID，则生成一个新的
		if requestID == "" {
			requestID = generateRequestID()
		}

		// 设置响应头
		c.Header("X-Request-ID", requestID)

		// 将请求ID存储到上下文中，供后续处理使用
		c.Set("request_id", requestID)

		// 继续处理请求
		c.Next()
	}
}

// generateRequestID 生成请求ID
// 使用时间戳纳秒作为唯一标识符
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
