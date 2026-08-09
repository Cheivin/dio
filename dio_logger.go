package dio

import (
	"context"

	"github.com/cheivin/di"
	"github.com/cheivin/dio-core"
)

type emptyLogger struct {
}

func (e emptyLogger) DebugMode(b bool) {
}

func (e emptyLogger) Debug(s string) {
}

func (e emptyLogger) Info(s string) {
}

func (e emptyLogger) Warn(s string) {
}

func (e emptyLogger) Fatal(err error) {
	panic(err)
}

// dioLogger 适配 di.Log 到 core.Log：di 的日志方法无 context 参数，
// 这里统一用 Background()（容器内部日志无请求上下文）。
type dioLogger struct {
	log core.Log
}

func newDiLogger(log core.Log) di.Log {
	return dioLogger{
		log: log.Named("[DIO]").Skip(0),
	}
}

func (d dioLogger) DebugMode(_ bool) {
	// di 的 DebugMode 对 dio 无意义：dio 的调试级别由 zap 的 log.debug 配置控制
}

func (d dioLogger) Debug(s string) {
	d.log.Debug(context.Background(), s)
}

func (d dioLogger) Info(s string) {
	d.log.Info(context.Background(), s)
}

func (d dioLogger) Warn(s string) {
	d.log.Warn(context.Background(), s)
}

func (d dioLogger) Fatal(err error) {
	d.log.Error(context.Background(), err.Error())
	panic(err)
}
