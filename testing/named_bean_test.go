package testing

import (
	"context"
	"testing"
	"time"

	"github.com/cheivin/dio"
)

type namedSvc struct{}

// TestProvideNamedBean 验证指定名称注册原型：名称生效且不再按类型推断。
// 注意：GetBean 基于实例，须在 Running 状态内断言（Serve 退出后 bean 已销毁）。
func TestProvideNamedBean(t *testing.T) {
	defer dio.Reset()
	dio.SetBanner("")
	dio.ProvideNamedBean("myService", namedSvc{})

	var found, inferredFound bool
	dio.OnStateChange(func(s dio.AppState) {
		if s == dio.Running {
			_, found = dio.GetBean("myService")
			_, inferredFound = dio.GetBean("namedSvc")
		}
	})
	runWithTimeout(t, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		dio.Run(ctx)
	})

	if !found {
		t.Fatal("GetBean(myService) should be found with explicit name")
	}
	if inferredFound {
		t.Fatal("bean should not be registered under inferred name namedSvc")
	}
}

// TestProvideMultiNamedBean 验证按 map 注册多个同名类型 bean：显式名称生效，不因推断重名 panic。
func TestProvideMultiNamedBean(t *testing.T) {
	defer dio.Reset()
	dio.SetBanner("")
	dio.ProvideMultiNamedBean(map[string]any{
		"svc1": namedSvc{},
		"svc2": namedSvc{},
	})

	var svc1, svc2 bool
	dio.OnStateChange(func(s dio.AppState) {
		if s == dio.Running {
			_, svc1 = dio.GetBean("svc1")
			_, svc2 = dio.GetBean("svc2")
		}
	})
	ok := false
	runWithTimeout(t, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Run should not panic with distinct names: %v", r)
			}
		}()
		dio.Run(ctx)
		ok = true
	})
	if !ok {
		t.Fatal("Run should complete")
	}
	if !svc1 {
		t.Fatal("GetBean(svc1) should be found")
	}
	if !svc2 {
		t.Fatal("GetBean(svc2) should be found")
	}
}
