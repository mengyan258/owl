package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	logContract "bit-labs.cn/owl/contract/log"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type OwlGormLogger struct {
	log                     logContract.Logger
	level                   logger.LogLevel
	slowThreshold           time.Duration
	ignoreRecordNotFoundErr bool
}

var _ logger.Interface = (*OwlGormLogger)(nil)

func NewOwlGormLogger(l logContract.Logger) *OwlGormLogger {
	return &OwlGormLogger{
		log:                     l,
		level:                   logger.Info,
		slowThreshold:           200 * time.Millisecond,
		ignoreRecordNotFoundErr: true,
	}
}

func (i *OwlGormLogger) LogMode(level logger.LogLevel) logger.Interface {
	n := *i
	n.level = level
	return &n
}

func (i *OwlGormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if i.level < logger.Info || i.log == nil {
		return
	}
	requestID := getRequestID(ctx)
	if requestID != "" {
		i.log.WithContext(ctx).Info("GORM", "requestId", requestID, "msg", fmt.Sprintf(msg, data...))
		return
	}
	i.log.WithContext(ctx).Info(fmt.Sprintf(msg, data...))
}

func (i *OwlGormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if i.level < logger.Warn || i.log == nil {
		return
	}
	requestID := getRequestID(ctx)
	if requestID != "" {
		i.log.WithContext(ctx).Warning("GORM", "requestId", requestID, "msg", fmt.Sprintf(msg, data...))
		return
	}
	i.log.WithContext(ctx).Warning(fmt.Sprintf(msg, data...))
}

func (i *OwlGormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if i.level < logger.Error || i.log == nil {
		return
	}
	requestID := getRequestID(ctx)
	if requestID != "" {
		i.log.WithContext(ctx).Error("GORM", "requestId", requestID, "msg", fmt.Sprintf(msg, data...))
		return
	}
	i.log.WithContext(ctx).Error(fmt.Sprintf(msg, data...))
}

func (i *OwlGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if i.level <= logger.Silent || i.log == nil {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	requestID := getRequestID(ctx)

	if err != nil {
		if i.ignoreRecordNotFoundErr && errors.Is(err, gorm.ErrRecordNotFound) {
			return
		}
		if i.level >= logger.Error {
			i.log.WithContext(ctx).Error("执行 SQL 失败", "requestId", requestID, "耗时", elapsed, "影响行数", rows, "sql:", sql, "错误:", err)
		}
		return
	}

	if i.slowThreshold > 0 && elapsed > i.slowThreshold {
		if i.level >= logger.Warn {
			i.log.WithContext(ctx).Warning("慢查询", "requestId", requestID, "耗时", elapsed, "影响行数", rows, "sql:", sql)
		}
		return
	}

	if i.level >= logger.Info {
		i.log.WithContext(ctx).Info("执行 SQL", "requestId", requestID, "耗时", elapsed, "影响行数", rows, "sql:", sql)
	}
}

func getRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v := ctx.Value("request_id")
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}
