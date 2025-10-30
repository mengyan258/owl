package log

import (
	"bit-labs.cn/owl/contract/log"
	"context"
	"github.com/golang-module/carbon"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"sync"
)

/*
日志初始化步骤：
1. 设置 log 写入的文件 (writer)
2. 设置为日志编码的方法（encoder）
3. 创建日志核心
*/

var _ log.Logger = (*FileImpl)(nil)

type FileImpl struct {
	options    *FileImplOptions
	l          *zap.SugaredLogger
	lock       sync.RWMutex
	preGetTime string
	ctx        context.Context
}

func (i *FileImpl) WithContext(ctx context.Context) log.Logger {
	i.ctx = ctx
	return i
}

type FileImplOptions struct {
	StorePath    string
	MaxSize      int
	MaxBackups   int
	MaxAge       int
	Compress     bool
	Level        zapcore.Level
	FileName     string
	dateFileName string // 计算出来，每天一个文件
}

var (
	defaultOptions = &FileImplOptions{
		StorePath:    "./storage",
		dateFileName: "0000-00-00",
		MaxSize:      50,
		MaxBackups:   100,
		MaxAge:       30,
		Compress:     true,
		Level:        zap.DebugLevel,
	}
)

var _ log.Logger = (*FileImpl)(nil)

func NewFileImpl(options *FileImplOptions) *FileImpl {
	if options == nil {
		options = defaultOptions
	}

	l := &FileImpl{
		options: options,
	}
	return l
}

func (i *FileImpl) Emergency(content ...any) {
	i.setLogger()
	i.l.DPanic(content)
}

func (i *FileImpl) Alert(content ...any) {
	i.setLogger()
	i.l.Error(content)
}

func (i *FileImpl) Critical(content ...any) {
	i.setLogger()
}

func (i *FileImpl) Error(content ...any) {
	i.setLogger()
	i.l.Error(content)
}

func (i *FileImpl) Warning(content ...any) {
	i.setLogger()
	i.l.Warn(content)
}

func (i *FileImpl) Notice(content ...any) {
	i.setLogger()
	i.l.Warn(content)
}

func (i *FileImpl) Info(content ...any) {
	i.setLogger()
	i.l.Info(content)
}

func (i *FileImpl) Debug(content ...any) {
	i.setLogger()
	i.l.Debug(content)
}

// 根据时间轮换，因为随时都有可能发生时间变化，调用日志方法之前需要先调用这个方法
func (i *FileImpl) setLogger() {

	i.lock.Lock()
	defer i.lock.Unlock()

	now := carbon.Now()
	nowDate := now.ToDateString()

	// 如果文件夹不存在，则递归创建
	if _, err := os.Stat(i.options.StorePath); os.IsNotExist(err) {
		err = os.MkdirAll(i.options.StorePath, 755)
		if err != nil {
			return
		}
	}

	// 重新换一个文件来记录日志
	if i.preGetTime != "" {
		if nowDate != i.preGetTime {
			i.l = nil
			i.preGetTime = nowDate
		}
	} else {
		i.preGetTime = nowDate
	}

	if i.l != nil {
		return
	}

	i.options.dateFileName = i.options.FileName + "-" + nowDate

	encoder := getEncoder()
	writeSyncer := getLogWriter(i.options)
	// 打印
	write2file := zapcore.AddSync(writeSyncer)
	write2stdout := zapcore.AddSync(os.Stdout)
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoder),
		zapcore.NewMultiWriteSyncer(write2file, write2stdout),
		i.options.Level,
	)

	zap.AddCaller() //添加将调用函数信息记录到日志中的功能。
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	i.l = logger.Sugar()
}

func getLogWriter(options *FileImplOptions) *lumberjack.Logger {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   options.StorePath + "/" + options.dateFileName + ".log", // 日志文件路径
		MaxSize:    options.MaxSize,                                         // 最大尺寸, M
		MaxBackups: options.MaxBackups,                                      // 备份数 在进行切割之前，日志文件的最大大小（以MB为单位）
		MaxAge:     options.MaxAge,                                          // 保留旧文件的最大天数
		Compress:   options.Compress,                                        // 是否压缩/归档旧文件
	}

	return lumberJackLogger
}

func getEncoder() zapcore.EncoderConfig {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000") // 修改时间编码器

	// 在日志文件中使用大写字母记录日志级别
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return encoderConfig
}
