package testing

import (
	"context"
	"testing"
	"time"

	"github.com/cheivin/dio"
)

// TestStartupDuration 验证启动耗时统计：未 Run 为 0，Run 后大于 0。
func TestStartupDuration(t *testing.T) {
	defer dio.Reset()
	if dur := dio.StartupDuration(); dur != 0 {
		t.Fatalf("before Run, StartupDuration should be 0, got %v", dur)
	}
	dio.SetBanner("") // 关闭 banner，保持测试输出干净
	runWithTimeout(t, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		dio.Run(ctx)
	})
	if dur := dio.StartupDuration(); dur <= 0 {
		t.Fatalf("after Run, StartupDuration should be > 0, got %v", dur)
	}
}
