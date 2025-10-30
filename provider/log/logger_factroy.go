package log

import (
	"bit-labs.cn/owl/contract/foundation"
)

type LoggerFactory struct {
	app foundation.Application
	opt *Options
}

type Options struct {
	Level      int  `json:"level"`
	MaxSize    int  `json:"max-size"`
	MaxBackups int  `json:"max-backups"`
	MaxAge     int  `json:"max-age"`
	Compress   bool `json:"compress"`
}
