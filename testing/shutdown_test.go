package testing

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/cheivin/dio"
)

// TestShutdownSequential 验证默认顺序模式：按注册倒序执行。
func TestShutdownSequential(t *testing.T) {
	defer dio.Reset()
	dio.SetBanner("")
	var order []string
	dio.OnShutdown(func() { order = append(order, "first") })
	dio.OnShutdown(func() { order = append(order, "second") })
	runWithTimeout(t, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		dio.Run(ctx)
	})
	want := []string{"second", "first"} // 后注册先执行
	if len(order) != len(want) {
		t.Fatalf("shutdown order = %v, want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("shutdown order = %v, want %v", order, want)
		}
	}
}

// TestShutdownParallel 验证并行模式：两个回调同时执行（互相等待不阻塞）。
func TestShutdownParallel(t *testing.T) {
	defer dio.Reset()
	dio.SetBanner("")
	dio.SetShutdownParallel(true)
	release := make(chan struct{})
	var mu sync.Mutex
	done := map[string]bool{}
	// 第一个回调（后执行）等待 release；并行模式下它先启动并阻塞等待，
	// 第二个回调关闭 release 后两者都能完成——顺序模式会死锁
	dio.OnShutdown(func() {
		<-release
		mu.Lock()
		done["blocking"] = true
		mu.Unlock()
	})
	dio.OnShutdown(func() {
		close(release)
		mu.Lock()
		done["releaser"] = true
		mu.Unlock()
	})
	ok := false
	runWithTimeout(t, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		dio.Run(ctx)
		ok = true
	})
	if !ok {
		t.Fatal("parallel shutdown should not deadlock")
	}
	mu.Lock()
	defer mu.Unlock()
	if !done["blocking"] || !done["releaser"] {
		t.Fatalf("both shutdown callbacks should run, got %v", done)
	}
}

// TestShutdownTimeoutSequential 验证顺序模式超时：超时后跳过剩余回调。
func TestShutdownTimeoutSequential(t *testing.T) {
	defer dio.Reset()
	dio.SetBanner("")
	dio.SetShutdownTimeout(300 * time.Millisecond)
	skipped := false
	// 后注册的先执行：第一个回调耗时 350ms 超过 300ms 时限，第二个回调应被跳过
	dio.OnShutdown(func() { skipped = true })
	dio.OnShutdown(func() { time.Sleep(350 * time.Millisecond) })
	ok := false
	runWithTimeout(t, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		dio.Run(ctx)
		ok = true
	})
	if !ok {
		t.Fatal("shutdown should complete after timeout")
	}
	if skipped {
		t.Fatal("remaining shutdown callback should be skipped after timeout")
	}
}

// TestShutdownTimeoutParallel 验证并行模式超时：超时后不再等待，Run 返回。
func TestShutdownTimeoutParallel(t *testing.T) {
	defer dio.Reset()
	dio.SetBanner("")
	dio.SetShutdownParallel(true)
	dio.SetShutdownTimeout(100 * time.Millisecond)
	ran := make(chan struct{})
	dio.OnShutdown(func() { time.Sleep(500 * time.Millisecond) })
	dio.OnShutdown(func() { close(ran) })
	ok := false
	runWithTimeout(t, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		dio.Run(ctx)
		ok = true
	})
	if !ok {
		t.Fatal("parallel shutdown should not block past timeout")
	}
	select {
	case <-ran:
	case <-time.After(1 * time.Second):
		t.Fatal("fast shutdown callback should run before timeout")
	}
	if dio.State() != dio.Stopped {
		t.Fatalf("state should be Stopped, got %s", dio.State())
	}
}
