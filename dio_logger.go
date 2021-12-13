package dio

import (
	"context"
	"fmt"
	"github.com/cheivin/di"
	"github.com/cheivin/dio-core/system"
	"go.uber.org/zap"
	"os"
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

func (e emptyLogger) Fatal(s string) {
	panic(s)
}

type dioLogger struct {
	ctx context.Context
	log *zap.Logger
}

func newDiLogger(ctx context.Context, log *system.Log) di.Log {
	fmt.Println(" ____    ______   _____      \n/\\  _`\\ /\\__  _\\ /\\  __`\\    \n\\ \\ \\/\\ \\/_/\\ \\/ \\ \\ \\/\\ \\   \n \\ \\ \\ \\ \\ \\ \\ \\  \\ \\ \\ \\ \\  \n  \\ \\ \\_\\ \\ \\_\\ \\__\\ \\ \\_\\ \\ \n   \\ \\____/ /\\_____\\\\ \\_____\\\n    \\/___/  \\/_____/ \\/_____/")
	return dioLogger{
		ctx: ctx,
		log: log.Logger().Named("[DIO]").WithOptions(zap.WithCaller(false)),
	}
}

func (d dioLogger) DebugMode(b bool) {
	if b {
		d.log = d.log.WithOptions(zap.Development())
	}
}

func (d dioLogger) Debug(s string) {
	d.log.Debug(s)
}

func (d dioLogger) Info(s string) {
	d.log.Info(s)
}

func (d dioLogger) Warn(s string) {
	d.log.Warn(s)
}

func (d dioLogger) Fatal(s string) {
	d.log.Error(s)
	os.Exit(1)
}
