package dio

import (
	"context"
	"fmt"
	"github.com/cheivin/di"
	"github.com/cheivin/dio-core"
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
	log core.Log
}

func newDiLogger(ctx context.Context, log core.Log) di.Log {
	fmt.Println(" ____    ______   _____      \n/\\  _`\\ /\\__  _\\ /\\  __`\\    \n\\ \\ \\/\\ \\/_/\\ \\/ \\ \\ \\/\\ \\   \n \\ \\ \\ \\ \\ \\ \\ \\  \\ \\ \\ \\ \\  \n  \\ \\ \\_\\ \\ \\_\\ \\__\\ \\ \\_\\ \\ \n   \\ \\____/ /\\_____\\\\ \\_____\\\n    \\/___/  \\/_____/ \\/_____/")
	return dioLogger{
		ctx: ctx,
		log: log.Named("[DIO]").Skip(0),
	}
}

func (d dioLogger) DebugMode(_ bool) {
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

func (d dioLogger) Fatal(s string) {
	d.log.Error(context.Background(), s)
	os.Exit(1)
}
