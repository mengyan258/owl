package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// SecurityConfig 安全配置
type SecurityConfig struct {
	ContentTypeNoSniff bool `yaml:"content-type-nosniff"`
	XssProtection      bool `yaml:"xss-protection"`
	FrameDeny          bool `yaml:"frame-deny"`
	Hsts               bool `yaml:"hsts"`
	HstsMaxAge         int  `yaml:"hsts-max-age"`
}

// Security 安全头中间件
// 添加各种安全相关的HTTP头，提高应用安全性
func Security(config SecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// X-Content-Type-Options: 防止浏览器进行MIME类型嗅探
		if config.ContentTypeNoSniff {
			c.Header("X-Content-Type-Options", "nosniff")
		}

		// X-XSS-Protection: 启用浏览器的XSS过滤器
		if config.XssProtection {
			c.Header("X-XSS-Protection", "1; mode=block")
		}

		// X-Frame-Options: 防止点击劫持攻击
		if config.FrameDeny {
			c.Header("X-Frame-Options", "DENY")
		}

		// Strict-Transport-Security: 强制使用HTTPS
		if config.Hsts {
			c.Header("Strict-Transport-Security", fmt.Sprintf("max-age=%d", config.HstsMaxAge))
		}

		// 继续处理请求
		c.Next()
	}
}

// DefaultSecurity 默认安全配置的中间件
func DefaultSecurity() gin.HandlerFunc {
	return Security(SecurityConfig{
		ContentTypeNoSniff: true,
		XssProtection:      true,
		FrameDeny:          true,
		Hsts:               false,    // 默认不启用HSTS，因为需要HTTPS环境
		HstsMaxAge:         31536000, // 1年
	})
}
