package dio

import (
	"context"
	"github.com/cheivin/dio-core"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path"
	"strings"
	"time"
)

type ZapLogger struct {
	traceName string // 会话追踪名称
	logger    *zap.SugaredLogger
}

func WrapZapLogger(logger *zap.Logger, opts ...zap.Option) core.Log {
	return &ZapLogger{logger: logger.WithOptions(opts...).Sugar()}
}

func NewZapLogger(l core.Property, opts ...zap.Option) (core.Log, error) {
	// 处理配置参数默认值
	if l.Name == "" {
		l.Name = "log"
	}
	if strings.Contains(l.Name, "@hostname") {
		hostname, _ := os.Hostname()
		l.Name = strings.ReplaceAll(l.Name, "@hostname", hostname)
	}
	if l.MaxAge <= 0 {
		l.MaxAge = 7
	}
	// 不输出文件的时候强制开启输出控制台
	if l.File == false && l.Std == false {
		l.Std = true
	}
	// 开始配置zap日志
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
	var cores []zapcore.Core
	// 输出到文件
	if l.File {
		cores = []zapcore.Core{}

		if infoWriter, err := getLogWriter(path.Join(l.Dir, l.Name)+".log", time.Duration(l.MaxAge)*time.Hour*24); err != nil {
			return nil, err
		} else {
			cores = append(cores, zapcore.NewCore(getLogEncoder(), zapcore.AddSync(infoWriter), levelEnable))
		}
		if errorWriter, err := getLogWriter(path.Join(l.Dir, l.Name)+"_error.log", time.Duration(l.MaxAge)*time.Hour*24); err != nil {
			return nil, err
		} else {
			cores = append(cores, zapcore.NewCore(getLogEncoder(), zapcore.AddSync(errorWriter), zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
				return lvl >= zapcore.ErrorLevel
			})))
		}
	}
	// 输出到控制台,
	if len(cores) == 0 || l.Std {
		cores = append(cores, zapcore.NewCore(getLogColorLevelEncoder(), zapcore.Lock(os.Stdout), levelEnable))
	}
	core := zapcore.NewTee(cores...)
	zapLogger := zap.New(core).WithOptions(options...)

	logger := WrapZapLogger(zapLogger, opts...).(*ZapLogger)
	logger.traceName = l.TraceName
	return logger, nil
}

func checkDir(filename string) error {
	dir := path.Dir(filename)
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func getLogWriter(filename string, maxAge time.Duration) (zapcore.WriteSyncer, error) {
	if err := checkDir(filename); err != nil {
		return nil, err
	}
	if writer, err := rotatelogs.New(
		strings.Replace(filename, ".log", "", -1)+"_%Y-%m-%d.log",
		rotatelogs.WithMaxAge(maxAge),
		rotatelogs.WithRotationTime(24*time.Hour),
	); err != nil {
		return nil, err
	} else {
		return zapcore.AddSync(writer), nil
	}
}

func getLogEncoder() zapcore.Encoder {
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

func getLogColorLevelEncoder() zapcore.Encoder {
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

func (l *ZapLogger) BeanName() string {
	return "log"
}

func (l *ZapLogger) Named(named string) (logger core.Log) {
	logger = WrapZapLogger(l.logger.Desugar().Named(named))
	logger.(*ZapLogger).traceName = l.traceName
	return
}

func (l *ZapLogger) Skip(skip int) (logger core.Log) {
	if skip <= 0 {
		logger = WrapZapLogger(l.logger.Desugar().WithOptions(zap.WithCaller(false)))
	} else {
		logger = WrapZapLogger(l.logger.Desugar().WithOptions(zap.WithCaller(true), zap.AddCallerSkip(skip)))
	}
	logger.(*ZapLogger).traceName = l.traceName
	return
}

func (l *ZapLogger) Logger() interface{} {
	return l.logger.Desugar()
}

func (l *ZapLogger) getTraceId(ctx context.Context) string {
	if l.traceName != "" {
		if id, ok := ctx.Value(l.traceName).(string); ok {
			return id
		}
	}
	return ""
}

func (l *ZapLogger) Trace(ctx context.Context) context.Context {
	if l.getTraceId(ctx) == "" {
		ctx = context.WithValue(ctx, l.traceName, core.UUID())
	}
	return ctx
}

func (l *ZapLogger) TraceWith(ctx context.Context, val any) context.Context {
	ctx = context.WithValue(ctx, l.traceName, val)
	return ctx
}

func (l *ZapLogger) map2slice(keyAndValues ...map[string]interface{}) (fields []interface{}) {
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

func (l *ZapLogger) Debug(ctx context.Context, msg string, keyAndValues ...interface{}) {
	l.logger.Named(l.getTraceId(ctx)).With(keyAndValues...).Debug(msg)
}

func (l *ZapLogger) Info(ctx context.Context, msg string, keyAndValues ...interface{}) {
	l.logger.Named(l.getTraceId(ctx)).With(keyAndValues...).Info(msg)
}

func (l *ZapLogger) Warn(ctx context.Context, msg string, keyAndValues ...interface{}) {
	l.logger.Named(l.getTraceId(ctx)).With(keyAndValues...).Warn(msg)
}

func (l *ZapLogger) Error(ctx context.Context, msg string, keyAndValues ...interface{}) {
	l.logger.Named(l.getTraceId(ctx)).With(keyAndValues...).Error(msg)
}

func (l *ZapLogger) Debugw(ctx context.Context, msg string, keyAndValues ...map[string]interface{}) {
	l.logger.Named(l.getTraceId(ctx)).With(l.map2slice(keyAndValues...)...).Debug(msg)
}

func (l *ZapLogger) Infow(ctx context.Context, msg string, keyAndValues ...map[string]interface{}) {
	l.logger.Named(l.getTraceId(ctx)).With(l.map2slice(keyAndValues...)...).Info(msg)
}

func (l *ZapLogger) Warnw(ctx context.Context, msg string, keyAndValues ...map[string]interface{}) {
	l.logger.Named(l.getTraceId(ctx)).With(l.map2slice(keyAndValues...)...).Warn(msg)
}

func (l *ZapLogger) Errorw(ctx context.Context, msg string, keyAndValues ...map[string]interface{}) {
	l.logger.Named(l.getTraceId(ctx)).With(l.map2slice(keyAndValues...)...).Error(msg)
}
