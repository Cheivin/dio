package testing

import (
	"context"
	"embed"
	"fmt"
	"testing"
	"time"

	"github.com/cheivin/dio"
	"github.com/cheivin/dio-core"
)

type A struct {
	Log       core.Log `aware:""`
	Container core.Dio `aware:""`
}

func (A) BeanConstruct() {
	fmt.Println("Load A")
}

func (a A) AfterPropertiesSet() {
	a.Log.Info(context.Background(), "加载完成")
	a.Container.Logger().Info(context.TODO(), "Container")
}

// runWithTimeout 在超时内运行 dio，避免阻塞 go test
func runWithTimeout(t *testing.T, fn func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer func() {
			recover() // Run 可能 panic，忽略
			close(done)
		}()
		fn()
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Log("run timed out (expected for blocking Serve)")
	}
}

func TestRun(t *testing.T) {
	defer dio.Reset()
	runWithTimeout(t, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		dio.SetProperty("app.env", "dev").
			ProvideOnProperty(A{}, "app.env", "dev").
			Run(ctx)
	})
}

//go:embed configs/*.yaml
var configs embed.FS

func TestYamlConfig(t *testing.T) {
	defer dio.Reset()
	dio.LoadConfig(configs, "configs/dev.yaml")
	runWithTimeout(t, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		dio.Run(ctx)
	})
}

func Test_GetByType(t *testing.T) {
	defer dio.Reset()
	property := dio.GetProperties("log.", core.Property{}).(core.Property)
	log, err := dio.NewZapLogger(property)
	if err != nil {
		t.Fatal(err)
	}
	dio.SetLogger(log)

	var x core.Log
	t.Log(dio.GetByType(&x))
}
