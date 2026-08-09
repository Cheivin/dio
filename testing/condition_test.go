package testing

import (
	"testing"

	"github.com/cheivin/dio"
	"github.com/cheivin/dio-core"
)

// TestOnProfile 验证 OnProfile 立即求值：profile 匹配时执行 fn，不匹配时不执行。
func TestOnProfile(t *testing.T) {
	defer dio.Reset()
	ran := ""
	dio.SetProfile("dev")
	dio.OnProfile("dev", func(d core.Dio) { ran = "dev" })
	dio.OnProfile("prod", func(d core.Dio) { ran = "prod" })
	if ran != "dev" {
		t.Fatalf("OnProfile result = %q, want dev", ran)
	}

	// 环境变量 APP_PROFILE 同样生效
	t.Setenv("APP_PROFILE", "test")
	dio.SetProfile("")
	dio.OnProfile("test", func(d core.Dio) { ran = "test" })
	if ran != "test" {
		t.Fatalf("OnProfile with APP_PROFILE result = %q, want test", ran)
	}
}

type condIface interface{ M() }

type condImplA struct{}

func (*condImplA) M() {}

type condImplB struct{}

// TestOnBeanType 验证 OnBeanType 立即求值：实例/原型/工厂 bean 均能命中，未注册不执行。
func TestOnBeanType(t *testing.T) {
	defer dio.Reset()
	ran := false

	// 未注册时不执行
	dio.OnBeanType(condImplA{}, func(d core.Dio) { ran = true })
	if ran {
		t.Fatal("OnBeanType should not run for unregistered type")
	}

	// 原型命中
	dio.Provide(condImplA{})
	dio.OnBeanType(condImplA{}, func(d core.Dio) { ran = true })
	if !ran {
		t.Fatal("OnBeanType should run for provided prototype")
	}

	// 接口命中（*condImplA 实现 condIface）
	ran = false
	dio.OnBeanType((*condIface)(nil), func(d core.Dio) { ran = true })
	if !ran {
		t.Fatal("OnBeanType should run for interface type")
	}

	// 实例命中
	ran = false
	dio.RegisterBean(&condImplB{})
	dio.OnBeanType(condImplB{}, func(d core.Dio) { ran = true })
	if !ran {
		t.Fatal("OnBeanType should run for registered instance")
	}

	// 工厂 bean 命中
	ran = false
	dio.ProvideFunc(func() *condImplB { return &condImplB{} })
	dio.OnBeanType(condImplB{}, func(d core.Dio) { ran = true })
	if !ran {
		t.Fatal("OnBeanType should run for factory bean")
	}
}
