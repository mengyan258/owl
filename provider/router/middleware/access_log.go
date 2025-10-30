package middleware

import (
	"github.com/gin-gonic/gin"
)

// AccessLogConfig 访问日志配置
type AccessLogConfig struct {
	Enabled   bool     `yaml:"enabled"`
	Format    string   `yaml:"format"`
	SkipPaths []string `yaml:"skip-paths"`
}

// AccessLog 访问日志中间件
// 记录HTTP请求的访问日志，支持跳过指定路径
func AccessLog(config AccessLogConfig) gin.HandlerFunc {
	if config.Enabled {
		return gin.Logger()
	}

	return func(c *gin.Context) {
		c.Next()
	}
}

// DefaultAccessLog 默认访问日志中间件
func DefaultAccessLog() gin.HandlerFunc {
	return AccessLog(AccessLogConfig{
		Enabled: true,
		Format:  "combined",
		SkipPaths: []string{
			"/health",
			"/metrics",
			"/favicon.ico",
		},
	})
}
