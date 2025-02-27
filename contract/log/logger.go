package log

import "bit-labs.cn/owl/contract"

// 定义日志级别常量
const (
	EMERGENCY Level = "emergency" // 系统无法使用，紧急情况
	ALERT     Level = "alert"     // 必须立即采取行动，例如：整个网站宕机、数据库挂起等，应发送短信或警报通知
	CRITICAL  Level = "critical"  // 临界条件，例如：应用组件不可用、严重异常
	ERROR     Level = "error"     // 运行时错误，不需要立即处理但通常应记录并监控
	WARNING   Level = "warning"   // 警告事件，非错误状态，例如：使用过时API、API使用不当等
	NOTICE    Level = "notice"    // 正常但重要的事件，值得关注的通知信息
	INFO      Level = "info"      // 一般信息性事件，例如：用户登录、SQL日志等
	DEBUG     Level = "debug"     // 详细的调试信息，用于开发和排错阶段
)

// Level 日志级别类型
type Level string

// Logger 接口定义了不同日志级别的方法
type Logger interface {
	contract.WithContext[Logger]
	// Emergency 输出紧急日志信息
	Emergency(content ...interface{})
	// Alert 输出警告日志信息，需要立即关注并采取行动
	Alert(content ...interface{})
	// Critical 输出临界条件日志信息
	Critical(content ...interface{})
	// Error 输出运行时错误日志信息
	Error(content ...interface{})
	// Warning 输出警告事件日志信息
	Warning(content ...interface{})
	// Notice 输出正常但重要的事件日志信息
	Notice(content ...interface{})
	// Info 输出一般信息事件日志
	Info(content ...interface{})
	// Debug 输出详细调试信息日志
	Debug(content ...interface{})
}

type Options struct {
	Level      string `json:"level"`
	MaxSize    int    `json:"max-size"`
	MaxBackups int    `json:"max-backups"`
	MaxAge     int    `json:"max-age"`
	Compress   bool   `json:"compress"`
}
