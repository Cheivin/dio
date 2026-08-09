package testing

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cheivin/dio"
)

type checkOK struct{}

func (checkOK) Health(ctx context.Context) error { return nil }

var errDbDown = errors.New("db down")

type checkFail struct{}

func (checkFail) Health(ctx context.Context) error { return errDbDown }

// checkSlow 依赖 ctx 取消：超时或被调用方 deadline 收紧时返回错误
type checkSlow struct{}

func (checkSlow) Health(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(1 * time.Second):
		return nil
	}
}

// runAndHealth 在容器 Running 状态下执行 health 检查并返回结果。
// Run 本身在独立 goroutine 中运行（runWithTimeout），健康检查通过 Running 状态回调触发。
func runAndHealth(t *testing.T) error {
	t.Helper()
	var healthErr error
	dio.OnStateChange(func(s dio.AppState) {
		if s == dio.Running {
			healthErr = dio.Health(context.Background())
		}
	})
	runWithTimeout(t, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		dio.Run(ctx)
	})
	return healthErr
}

// TestHealthNotReady 验证未就绪（未 Run 或非 Running）时 Health 直接返回错误。
func TestHealthNotReady(t *testing.T) {
	defer dio.Reset()
	if err := dio.Health(context.Background()); err == nil {
		t.Fatal("Health before Run should return error")
	}
}

// TestHealthOK 验证全部检查通过时 Health 返回 nil。
func TestHealthOK(t *testing.T) {
	defer dio.Reset()
	dio.SetBanner("")
	dio.Provide(checkOK{})
	if err := runAndHealth(t); err != nil {
		t.Fatalf("Health should be nil, got %v", err)
	}
}

// TestHealthFail 验证单个检查失败时 Health 返回聚合错误（可用 errors.Is 判断）。
func TestHealthFail(t *testing.T) {
	defer dio.Reset()
	dio.SetBanner("")
	dio.Provide(checkOK{}, checkFail{})
	err := runAndHealth(t)
	if err == nil {
		t.Fatal("Health should return error when a checker fails")
	}
	if !errors.Is(err, errDbDown) {
		t.Fatalf("Health error should contain errDbDown, got %v", err)
	}
}

// TestHealthTimeout 验证 ctx 带 deadline 时每个检查受其约束，超时返回错误。
func TestHealthTimeout(t *testing.T) {
	defer dio.Reset()
	dio.SetBanner("")
	dio.Provide(checkSlow{})
	var healthErr error
	dio.OnStateChange(func(s dio.AppState) {
		if s == dio.Running {
			// 100ms 的 deadline 应被传递给每个检查
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			healthErr = dio.Health(ctx)
		}
	})
	runWithTimeout(t, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		dio.Run(ctx)
	})
	if healthErr == nil {
		t.Fatal("Health should return error when check exceeds ctx deadline")
	}
	if !errors.Is(healthErr, context.DeadlineExceeded) {
		t.Fatalf("Health error should contain DeadlineExceeded, got %v", healthErr)
	}
}
