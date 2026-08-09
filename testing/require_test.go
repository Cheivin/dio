package testing

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cheivin/dio"
)

// TestRequirePropertiesMissing 验证必填配置缺失时 Run panic（ErrMissingProperty）且状态置为 Failed。
func TestRequirePropertiesMissing(t *testing.T) {
	defer dio.Reset()
	dio.SetBanner("")
	dio.RequireProperties("app.port", "db.host")

	var panicErr error
	runWithTimeout(t, func() {
		defer func() {
			if r := recover(); r != nil {
				if err, ok := r.(error); ok {
					panicErr = err
				}
			}
		}()
		dio.Run(context.Background())
	})
	if panicErr == nil {
		t.Fatal("Run should panic with ErrMissingProperty")
	}
	if !errors.Is(panicErr, dio.ErrMissingProperty) {
		t.Fatalf("panic error should be ErrMissingProperty, got %q", panicErr)
	}
	if dio.State() != dio.Failed {
		t.Fatalf("state should be Failed, got %s", dio.State())
	}

	// 补上缺失配置后重试成功
	dio.SetProperty("app.port", 8080)
	dio.SetProperty("db.host", "localhost")
	ok := false
	runWithTimeout(t, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("retry Run should not panic: %v", r)
			}
		}()
		dio.Run(ctx)
		ok = true
	})
	if !ok {
		t.Fatal("retry Run should complete")
	}
	if dio.State() != dio.Stopped {
		t.Fatalf("state should be Stopped, got %s", dio.State())
	}
}

// TestRequirePropertiesPresent 验证必填配置齐全时正常启动。
func TestRequirePropertiesPresent(t *testing.T) {
	defer dio.Reset()
	dio.SetBanner("")
	dio.SetProperty("app.port", 8080)
	dio.RequireProperties("app.port")
	ok := false
	runWithTimeout(t, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Run should not panic: %v", r)
			}
		}()
		dio.Run(ctx)
		ok = true
	})
	if !ok {
		t.Fatal("Run should complete")
	}
}
