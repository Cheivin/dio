package testing

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/cheivin/dio"
)

// TestStateTransition 验证状态机完整推进：Pending→Starting→Running→Stopping→Stopped，
// 且 OnStateChange 回调倒序执行、Ready() 仅在 Running 时为 true。
func TestStateTransition(t *testing.T) {
	defer dio.Reset()
	if dio.State() != dio.Pending {
		t.Fatalf("before Run, state should be Pending, got %s", dio.State())
	}
	if dio.Ready() {
		t.Fatal("before Run, Ready() should be false")
	}

	var states []dio.AppState
	// 注册两个回调验证倒序执行：state2 后注册，应先于 state1 收到通知
	var order []string
	dio.OnStateChange(func(s dio.AppState) {
		states = append(states, s)
		order = append(order, "state1")
	})
	dio.OnStateChange(func(s dio.AppState) {
		order = append(order, "state2")
		if s == dio.Running {
			if !dio.Ready() {
				t.Error("on Running, Ready() should be true")
			}
		} else if dio.Ready() {
			t.Errorf("on %s, Ready() should be false", s)
		}
	})

	runWithTimeout(t, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		dio.Run(ctx)
	})

	want := []dio.AppState{dio.Starting, dio.Running, dio.Stopping, dio.Stopped}
	if len(states) != len(want) {
		t.Fatalf("state transitions = %v, want %v", states, want)
	}
	for i := range want {
		if states[i] != want[i] {
			t.Fatalf("state transitions = %v, want %v", states, want)
		}
	}
	if dio.State() != dio.Stopped {
		t.Fatalf("after Run, state should be Stopped, got %s", dio.State())
	}
	// 倒序执行：每次状态推进都是后注册的 state2 先于 state1
	for i := 0; i < len(order); i += 2 {
		if order[i] != "state2" || order[i+1] != "state1" {
			t.Fatalf("OnStateChange callbacks not in reverse order: %v", order)
		}
	}
}

// TestStateFailed 验证 Run 启动失败时状态置为 Failed，修正后可重试（Failed→Starting 例外）。
func TestStateFailed(t *testing.T) {
	defer dio.Reset()
	// log.dir 的父路径是普通文件（/dev/null），MkdirAll 必然失败，NewZapLogger 返回错误
	dio.SetProperty("log.dir", "/dev/null/sub")

	runWithTimeout(t, func() {
		dio.Run(context.Background())
	})
	if dio.State() != dio.Failed {
		t.Fatalf("after failed Run, state should be Failed, got %s", dio.State())
	}
	if dio.Ready() {
		t.Fatal("after failed Run, Ready() should be false")
	}

	// 修正配置后重试成功（日志创建阶段失败不污染 di 容器）
	dio.SetProperty("log.dir", "./logs")
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
		t.Fatalf("after retry Run, state should be Stopped, got %s", dio.State())
	}
}

// TestStateString 验证状态名的可读输出。
func TestStateString(t *testing.T) {
	states := []struct {
		s    dio.AppState
		want string
	}{
		{dio.Pending, "Pending"},
		{dio.Starting, "Starting"},
		{dio.Running, "Running"},
		{dio.Stopping, "Stopping"},
		{dio.Stopped, "Stopped"},
		{dio.Failed, "Failed"},
	}
	for _, st := range states {
		if got := st.s.String(); got != st.want {
			t.Errorf("AppState(%d).String() = %q, want %q", st.s, st.want, got)
		}
	}
	if s := dio.AppState(99).String(); !strings.HasPrefix(s, "AppState(") {
		t.Errorf("invalid AppState String() = %q, want prefix AppState(", s)
	}
}
