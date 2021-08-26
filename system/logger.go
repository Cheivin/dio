package system

import (
	"context"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path"
	"strings"
	"time"
)

type Log struct {
	logger    *zap.SugaredLogger
	Name      string `value:"log.name"`
	Dir       string `value:"log.dir"`
	MaxAge    int    `value:"log.max-age"` // 存活日期，单位天
	DebugMode bool   `value:"log.debug"`
	Std       bool   `value:"log.std"`
	TraceName string `value:"log.trace-Name"` // 会话追踪名称
}

// BeanConstruct 初始化实例时，创建gin框架
func (l *Log) BeanConstruct() {
	// 处理配置参数默认值
	if l.Name == "" {
		l.Name = "log"
	}
	if l.MaxAge <= 0 {
		l.MaxAge = 7
	}

	// 开始配置zap日志
	infoWriter := l.getLogWriter(path.Join(l.Dir, l.Name) + ".log")
	errorWriter := l.getLogWriter(path.Join(l.Dir, l.Name) + "_error.log")

	var levelEnable zap.LevelEnablerFunc
	var options []zap.Option
	if l.DebugMode {
		levelEnable = func(lvl zapcore.Level) bool {
			return lvl >= zapcore.DebugLevel
		}
		options = append(options, zap.AddCaller(), zap.AddCallerSkip(1))
	} else {
		levelEnable = func(lvl zapcore.Level) bool {
			return lvl >= zapcore.InfoLevel
		}
	}
	cores := []zapcore.Core{
		zapcore.NewCore(l.getLogEncoder(), zapcore.AddSync(infoWriter), levelEnable),
		zapcore.NewCore(l.getLogEncoder(), zapcore.AddSync(errorWriter), zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.ErrorLevel
		})),
	}
	if l.Std {
		cores = append(cores, zapcore.NewCore(l.getLogColorLevelEncoder(), zapcore.Lock(os.Stdout), zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.DebugLevel
		})))
	}
	core := zapcore.NewTee(cores...)
	zapLogger := zap.New(core).WithOptions(options...)
	l.logger = zapLogger.Sugar()
}

func (l *Log) getLogEncoder() zapcore.Encoder {
	return zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:       "ts",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.CapitalLevelEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		},
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})
}

func (l *Log) Skip(skip int) *Log {
	logger := l.logger.Desugar().WithOptions(zap.AddCallerSkip(skip)).Sugar()
	return &Log{
		logger:    logger,
		Name:      l.Name,
		Dir:       l.Dir,
		MaxAge:    l.MaxAge,
		DebugMode: l.DebugMode,
		Std:       l.Std,
		TraceName: l.TraceName,
	}
}

func (l *Log) getLogColorLevelEncoder() zapcore.Encoder {
	return zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:       "ts",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.CapitalColorLevelEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		},
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})
}

func (l *Log) getLogWriter(filename string) zapcore.WriteSyncer {
	l.checkDir(filename)
	writer, err := rotatelogs.New(
		strings.Replace(filename, ".log", "", -1)+"_%Y-%m-%d.log",
		rotatelogs.WithMaxAge(time.Duration(l.MaxAge)*time.Hour*24),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		panic(err)
	}
	return zapcore.AddSync(writer)
}

func (l *Log) checkDir(filename string) {
	dir := path.Dir(filename)
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
}

func (l *Log) map2slice(keyAndValues ...map[string]interface{}) (fields []interface{}) {
	if len(keyAndValues) == 0 {
		return nil
	}
	for _, keyAndValue := range keyAndValues {
		for key, value := range keyAndValue {
			fields = append(fields, key, value)
		}
	}
	return
}

func (l *Log) getTraceId(ctx context.Context) string {
	if l.TraceName != "" {
		if id, ok := ctx.Value(l.TraceName).(string); ok {
			return id
		}
	}
	return ""
}

func (l *Log) Debug(ctx context.Context, msg string, keyAndValues ...interface{}) {
	l.logger.Named(l.getTraceId(ctx)).With(keyAndValues...).Debug(msg)
}

func (l *Log) Info(ctx context.Context, msg string, keyAndValues ...interface{}) {
	l.logger.Named(l.getTraceId(ctx)).With(keyAndValues...).Info(msg)
}

func (l *Log) Warn(ctx context.Context, msg string, keyAndValues ...interface{}) {
	l.logger.Named(l.getTraceId(ctx)).With(keyAndValues...).Warn(msg)
}

func (l *Log) Error(ctx context.Context, msg string, keyAndValues ...interface{}) {
	l.logger.Named(l.getTraceId(ctx)).With(keyAndValues...).Error(msg)
}

func (l *Log) Debugw(ctx context.Context, msg string, keyAndValues ...map[string]interface{}) {
	l.logger.Named(l.getTraceId(ctx)).With(l.map2slice(keyAndValues...)...).Debug(msg)
}

func (l *Log) Infow(ctx context.Context, msg string, keyAndValues ...map[string]interface{}) {
	l.logger.Named(l.getTraceId(ctx)).With(l.map2slice(keyAndValues...)...).Info(msg)
}

func (l *Log) Warnw(ctx context.Context, msg string, keyAndValues ...map[string]interface{}) {
	l.logger.Named(l.getTraceId(ctx)).With(l.map2slice(keyAndValues...)...).Warn(msg)
}

func (l *Log) Errorw(ctx context.Context, msg string, keyAndValues ...map[string]interface{}) {
	l.logger.Named(l.getTraceId(ctx)).With(l.map2slice(keyAndValues...)...).Error(msg)
}
