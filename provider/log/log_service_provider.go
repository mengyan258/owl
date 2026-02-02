package log

import (
	_ "embed"

	"bit-labs.cn/owl/contract/foundation"
	logContract "bit-labs.cn/owl/contract/log"
	"bit-labs.cn/owl/provider/conf"
	"go.uber.org/zap/zapcore"
)

var _ foundation.ServiceProvider = (*LogServiceProvider)(nil)

type LogServiceProvider struct {
	app foundation.Application
}

func (i *LogServiceProvider) Description() string {
	return "日志系统与文件日志输出"
}

type option struct {
	Level      int    `json:"level"`
	FileName   string `json:"file-name"`
	MaxSize    int    `json:"max-size"`
	MaxBackups int    `json:"max-backups"`
	MaxAge     int    `json:"max-age"`
}

func (i *LogServiceProvider) Register() {

	i.app.Register(func(c *conf.Configure) logContract.Logger {
		var cfg option

		if err := c.GetConfig("log", &cfg); err != nil {
			panic(err)
		}

		return NewFileImpl(&FileImplOptions{
			StorePath:  i.app.GetStoragePath(),
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Level:      zapcore.Level(cfg.Level),
			FileName:   cfg.FileName,
		})
	})
}

func (i *LogServiceProvider) Boot() {

}

//go:embed log.yaml
var logConf string

func (i *LogServiceProvider) Conf() map[string]string {
	return map[string]string{
		"log.yaml": logConf,
	}
}
