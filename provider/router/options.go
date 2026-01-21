package router

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

type TLSConfig struct {
	Enabled           bool   `json:"enabled"`
	CertFile          string `json:"cert-file"`
	KeyFile           string `json:"key-file"`
	ClientCAFile      string `json:"client-ca-file"`
	RequireClientCert bool   `json:"require-client-cert"`
	MinVersion        string `json:"min-version"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host           string    `json:"host"`
	Port           int       `json:"port"`
	ReadTimeout    int       `json:"read-timeout"`
	WriteTimeout   int       `json:"write-timeout"`
	IdleTimeout    int       `json:"idle-timeout"`
	MaxHeaderBytes int       `json:"max-header-bytes"`
	TLS            TLSConfig `json:"tls"`
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
