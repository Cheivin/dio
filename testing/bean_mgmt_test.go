package testing

import (
	"context"
	"testing"
	"time"

	"github.com/cheivin/dio"
)

type mgmtDep struct{}

type mgmtService struct {
	Dep *mgmtDep `aware:""`
}

// TestBeanManagement 验证 Bean 管理 API：GetBeanNames/DescribeBean/GetBeanDependencies。
func TestBeanManagement(t *testing.T) {
	defer dio.Reset()
	dio.SetBanner("")
	dio.Provide(mgmtService{})
	dio.Provide(mgmtDep{})
	// ProvideFunc 直接进入 di，Run 前即可描述
	dio.ProvideFunc(func() *mgmtDep { return &mgmtDep{} })

	// 工厂 bean 立即可见（定义在 di）
	names := dio.GetBeanNames()
	found := false
	for _, name := range names {
		if name == "mgmtDep" {
			found = true
		}
	}
	if !found {
		t.Fatalf("GetBeanNames should contain mgmtDep, got %v", names)
	}

	// 原型 bean 需 Run 后才同步进 di
	runWithTimeout(t, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		dio.Run(ctx)
	})

	names = dio.GetBeanNames()
	if len(names) == 0 {
		t.Fatal("GetBeanNames after Run should not be empty")
	}

	desc, ok := dio.DescribeBean("mgmtService")
	if !ok {
		t.Fatal("DescribeBean should find mgmtService after Run")
	}
	if desc.Name != "mgmtService" || len(desc.Dependencies) != 1 || desc.Dependencies[0].Name != "mgmtDep" {
		t.Fatalf("unexpected description: %+v", desc)
	}

	deps, ok := dio.GetBeanDependencies("mgmtService")
	if !ok || len(deps) != 1 || deps[0] != "mgmtDep" {
		t.Fatalf("GetBeanDependencies(mgmtService) = %v, %v, want [mgmtDep], true", deps, ok)
	}

	// 不存在的 bean
	if _, ok := dio.DescribeBean("notExist"); ok {
		t.Fatal("DescribeBean should return ok=false for missing bean")
	}
	if _, ok := dio.GetBeanDependencies("notExist"); ok {
		t.Fatal("GetBeanDependencies should return ok=false for missing bean")
	}
}
