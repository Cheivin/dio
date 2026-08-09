package dio

import (
	"context"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/cheivin/dio-core"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Disposable 可释放资源的日志组件（如 ZapLogger 的 flush/关闭）。
// Run 启动失败时，容器会调用 Close 释放已创建的资源。
type Disposable interface {
	Close() error
}

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
		options = append(options, zap.AddCaller(), zap.AddCallerSkip(2))
	} else {
		levelEnable = func(lvl zapcore.Level) bool {
			return lvl >= zapcore.InfoLevel
		}
	}
	var cores []zapcore.Core
	// 输出到文件
	if l.File {
		cores = []zapcore.Core{}

		// 逐个创建 writer；仅创建失败时关闭已创建的，避免文件句柄泄漏。
		// 成功路径不关闭（writer 由 ZapLogger 生命周期管理）。
		var writers []io.Closer
		ok := false
		defer func() {
			if !ok {
				for _, w := range writers {
					_ = w.Close()
				}
			}
		}()
		if infoWriter, err := getLogWriter(path.Join(l.Dir, l.Name)+".log", time.Duration(l.MaxAge)*time.Hour*24); err != nil {
			return nil, err
		} else {
			writers = append(writers, infoWriter)
			cores = append(cores, zapcore.NewCore(getLogEncoder(false), zapcore.AddSync(infoWriter), levelEnable))
		}
		if errorWriter, err := getLogWriter(path.Join(l.Dir, l.Name)+"_error.log", time.Duration(l.MaxAge)*time.Hour*24); err != nil {
			return nil, err
		} else {
			writers = append(writers, errorWriter)
			cores = append(cores, zapcore.NewCore(getLogEncoder(false), zapcore.AddSync(errorWriter), zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
				return lvl >= zapcore.ErrorLevel
			})))
		}
		ok = true
	}
	// 输出到控制台,
	if len(cores) == 0 || l.Std {
		cores = append(cores, zapcore.NewCore(getLogEncoder(true), zapcore.Lock(os.Stdout), levelEnable))
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

func getLogWriter(filename string, maxAge time.Duration) (io.WriteCloser, error) {
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
		return writer, nil
	}
}

// getLogEncoder 构建控制台编码器。color 为 true 时 level 输出带颜色（用于控制台）。
func getLogEncoder(color bool) zapcore.Encoder {
	encodeLevel := zapcore.LevelEncoder(zapcore.CapitalLevelEncoder)
	if color {
		encodeLevel = zapcore.CapitalColorLevelEncoder
	}
	return zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:       "ts",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   encodeLevel,
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
		// +1 补偿 log helper 的额外栈帧，保持 caller 与 P2-2 前一致
		logger = WrapZapLogger(l.logger.Desugar().WithOptions(zap.WithCaller(true), zap.AddCallerSkip(skip+1)))
	}
	logger.(*ZapLogger).traceName = l.traceName
	return
}

func (l *ZapLogger) Logger() any {
	return l.logger.Desugar()
}

// Close flush 所有日志输出（文件/控制台）。
// 实现 dio.Disposable 接口，供容器在 Run 启动失败时清理资源。
func (l *ZapLogger) Close() error {
	return l.logger.Desugar().Sync()
}

// Destroy 实现 di.Disposable 接口。
// ZapLogger 作为 "log" bean 注册进容器，容器销毁（Serve 退出）时会调用它完成日志 flush。
func (l *ZapLogger) Destroy() {
	_ = l.Close()
}

// traceContextKey 是 traceId 在 context 中的 key 类型。
// 使用自定义类型而非裸 string，避免与其他库（gin/grpc 等）的 context key 冲突，
// 并符合 Go 的 context key 约定（staticcheck SA1029）。
type traceContextKey struct{ name string }

func (l *ZapLogger) getTraceId(ctx context.Context) string {
	if l.traceName != "" {
		if id, ok := ctx.Value(traceContextKey{name: l.traceName}).(string); ok {
			return id
		}
	}
	return ""
}

func (l *ZapLogger) Trace(ctx context.Context) context.Context {
	if l.traceName != "" && l.getTraceId(ctx) == "" {
		ctx = context.WithValue(ctx, traceContextKey{name: l.traceName}, core.UUID())
	}
	return ctx
}

func (l *ZapLogger) TraceWith(ctx context.Context, val any) context.Context {
	// 非 string 值无法作为 traceId 输出，忽略之
	if s, ok := val.(string); ok {
		ctx = context.WithValue(ctx, traceContextKey{name: l.traceName}, s)
	}
	return ctx
}

func (l *ZapLogger) map2slice(keyAndValues ...map[string]any) (fields []any) {
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

// log 内部统一出口：带上 traceId 的 named logger 输出。
func (l *ZapLogger) log(ctx context.Context, lvl zapcore.Level, msg string, fields ...any) {
	logger := l.logger.Named(l.getTraceId(ctx))
	if len(fields) > 0 {
		logger = logger.With(fields...)
	}
	logger.Log(lvl, msg)
}

func (l *ZapLogger) Debug(ctx context.Context, msg string, keyAndValues ...any) {
	l.log(ctx, zapcore.DebugLevel, msg, keyAndValues...)
}

func (l *ZapLogger) Info(ctx context.Context, msg string, keyAndValues ...any) {
	l.log(ctx, zapcore.InfoLevel, msg, keyAndValues...)
}

func (l *ZapLogger) Warn(ctx context.Context, msg string, keyAndValues ...any) {
	l.log(ctx, zapcore.WarnLevel, msg, keyAndValues...)
}

func (l *ZapLogger) Error(ctx context.Context, msg string, keyAndValues ...any) {
	l.log(ctx, zapcore.ErrorLevel, msg, keyAndValues...)
}

func (l *ZapLogger) Debugw(ctx context.Context, msg string, keyAndValues ...map[string]any) {
	l.log(ctx, zapcore.DebugLevel, msg, l.map2slice(keyAndValues...)...)
}

func (l *ZapLogger) Infow(ctx context.Context, msg string, keyAndValues ...map[string]any) {
	l.log(ctx, zapcore.InfoLevel, msg, l.map2slice(keyAndValues...)...)
}

func (l *ZapLogger) Warnw(ctx context.Context, msg string, keyAndValues ...map[string]any) {
	l.log(ctx, zapcore.WarnLevel, msg, l.map2slice(keyAndValues...)...)
}

func (l *ZapLogger) Errorw(ctx context.Context, msg string, keyAndValues ...map[string]any) {
	l.log(ctx, zapcore.ErrorLevel, msg, l.map2slice(keyAndValues...)...)
}
