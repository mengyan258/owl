package router

import (
	"bit-labs.cn/owl/contract/foundation"
	"bit-labs.cn/owl/provider/conf"
	"bit-labs.cn/owl/provider/router/middleware"
	_ "embed"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

var _ foundation.ServiceProvider = (*RouterServiceProvider)(nil)

type RouterServiceProvider struct {
	app foundation.Application
}

// RouterOptions 路由配置选项
type RouterOptions struct {
	Mode       string           `json:"mode"`
	Server     ServerConfig     `json:"server"`
	Middleware MiddlewareConfig `json:"middleware"`
	Cors       CorsConfig       `json:"cors"`
	RateLimit  RateLimitConfig  `json:"rate-limit"`
	Static     StaticConfig     `json:"static"`
	Template   TemplateConfig   `json:"template"`
	Security   SecurityConfig   `json:"security"`
	Log        LogConfig        `json:"log"`
	Health     HealthConfig     `json:"health"`
	Metrics    MetricsConfig    `json:"metrics"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host           string `json:"host"`
	Port           int    `json:"port"`
	ReadTimeout    int    `json:"read-timeout"`
	WriteTimeout   int    `json:"write-timeout"`
	IdleTimeout    int    `json:"idle-timeout"`
	MaxHeaderBytes int    `json:"max-header-bytes"`
}

// MiddlewareConfig 中间件配置
type MiddlewareConfig struct {
	Recovery  bool `json:"recovery"`
	Logger    bool `json:"logger"`
	Cors      bool `json:"cors"`
	RequestID bool `json:"request-id"`
	RateLimit bool `json:"rate-limit"`
}

// CorsConfig CORS 配置
type CorsConfig struct {
	AllowedOrigins   []string `json:"allowed-origins"`
	AllowedMethods   []string `json:"allowed-methods"`
	AllowedHeaders   []string `json:"allowed-headers"`
	ExposedHeaders   []string `json:"exposed-headers"`
	AllowCredentials bool     `json:"allow-credentials"`
	MaxAge           int      `json:"max-age"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	RequestsPerSecond int    `json:"requests-per-second"`
	Burst             int    `json:"burst"`
	KeyGenerator      string `json:"key-generator"`
}

// StaticConfig 静态文件配置
type StaticConfig struct {
	Enabled       bool   `json:"enabled"`
	Path          string `json:"path"`
	Root          string `json:"root"`
	ListDirectory bool   `json:"list-directory"`
}

// TemplateConfig 模板配置
type TemplateConfig struct {
	Enabled bool   `json:"enabled"`
	Pattern string `json:"pattern"`
	Delims  struct {
		Left  string `json:"left"`
		Right string `json:"right"`
	} `json:"delims"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	HttpsRedirect      bool `json:"https-redirect"`
	Hsts               bool `json:"hsts"`
	HstsMaxAge         int  `json:"hsts-max-age"`
	ContentTypeNoSniff bool `json:"content-type-nosniff"`
	XssProtection      bool `json:"xss-protection"`
	FrameDeny          bool `json:"frame-deny"`
}

// LogConfig 日志配置
type LogConfig struct {
	AccessLog       bool     `json:"access-log"`
	AccessLogFormat string   `json:"access-log-format"`
	SkipPaths       []string `json:"skip-paths"`
}

// HealthConfig 健康检查配置
type HealthConfig struct {
	Enabled bool   `json:"enabled"`
	Path    string `json:"path"`
}

// MetricsConfig 指标配置
type MetricsConfig struct {
	Enabled bool   `json:"enabled"`
	Path    string `json:"path"`
}

func (i *RouterServiceProvider) Register() {
	i.app.Register(func(c *conf.Configure) *gin.Engine {
		var opt RouterOptions
		err := c.GetConfig("router", &opt)
		if err != nil {
			panic(err)
		}

		return InitRouter(&opt)
	})
}

func (i *RouterServiceProvider) Boot() {
	// 路由服务启动时的初始化逻辑
}

//go:embed router.yaml
var routerYaml string

func (i *RouterServiceProvider) GenerateConf() map[string]string {
	return map[string]string{
		"router.yaml": routerYaml,
	}
}

// InitRouter 初始化路由引擎
func InitRouter(opt *RouterOptions) *gin.Engine {
	// 设置 Gin 模式
	gin.SetMode(opt.Mode)

	// 创建 Gin 引擎
	var engine *gin.Engine
	if opt.Middleware.Recovery || opt.Middleware.Logger {
		engine = gin.Default()
	} else {
		engine = gin.New()
	}

	// 配置中间件
	setupMiddleware(engine, opt)

	// 配置静态文件
	if opt.Static.Enabled {
		if opt.Static.ListDirectory {
			engine.StaticFS(opt.Static.Path, http.Dir(opt.Static.Root))
		} else {
			engine.Static(opt.Static.Path, opt.Static.Root)
		}
	}

	// 配置模板
	if opt.Template.Enabled {
		engine.LoadHTMLGlob(opt.Template.Pattern)
		if opt.Template.Delims.Left != "" && opt.Template.Delims.Right != "" {
			engine.Delims(opt.Template.Delims.Left, opt.Template.Delims.Right)
		}
	}

	// 配置健康检查
	if opt.Health.Enabled {
		engine.GET(opt.Health.Path, func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status": "ok",
			})
		})
	}

	// 配置指标端点
	if opt.Metrics.Enabled {
		engine.GET(opt.Metrics.Path, func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status": "ok",
			})
		})
	}

	return engine
}

// setupMiddleware 设置中间件
func setupMiddleware(engine *gin.Engine, opt *RouterOptions) {
	// CORS 中间件
	if opt.Middleware.Cors {
		config := cors.Config{
			AllowOrigins:     opt.Cors.AllowedOrigins,
			AllowMethods:     opt.Cors.AllowedMethods,
			AllowHeaders:     opt.Cors.AllowedHeaders,
			ExposeHeaders:    opt.Cors.ExposedHeaders,
			AllowCredentials: opt.Cors.AllowCredentials,
			MaxAge:           time.Duration(opt.Cors.MaxAge) * time.Second,
		}
		if len(opt.Cors.AllowedOrigins) == 0 {
			// 如果没有配置，默认允许所有
			config.AllowAllOrigins = true
		}
		engine.Use(cors.New(config))
	}

	// 请求 ID 中间件
	if opt.Middleware.RequestID {
		engine.Use(middleware.RequestID())
	}

	// 安全头中间件
	securityConfig := middleware.SecurityConfig{
		ContentTypeNoSniff: opt.Security.ContentTypeNoSniff,
		XssProtection:      opt.Security.XssProtection,
		FrameDeny:          opt.Security.FrameDeny,
		Hsts:               opt.Security.Hsts,
		HstsMaxAge:         opt.Security.HstsMaxAge,
	}
	engine.Use(middleware.Security(securityConfig))
}
