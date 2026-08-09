---
layout: default
title: 快速开始
nav_order: 3
parent: 入门
---

# 快速开始

一个完整的 dio 应用：配置加载 → 注册 bean → 启动 → 优雅退出。

## 最小应用

```go
package main

import (
	"context"

	"github.com/cheivin/dio"
)

type Greeter struct{}

func (g Greeter) Greet() string { return "hello dio" }

func main() {
	dio.SetProperty("app.port", 8080).
		Provide(Greeter{}).
		Run(context.Background())
}
```

`Run` 会阻塞等待退出信号（SIGINT/SIGTERM）或 ctx 取消，退出时自动销毁所有 bean 并执行停机回调。

## 带依赖注入的完整示例

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cheivin/dio"
	"github.com/cheivin/dio-core"
)

type DB struct{}

type UserService struct {
	Log       core.Log `aware:""` // 注入日志
	DB        *DB      `aware:""` // 注入其他 bean
	AppEnv    string   `value:"app.env"`
}

func (u *UserService) AfterPropertiesSet() {
	u.Log.Info(context.Background(), "UserService 初始化完成", "env", u.AppEnv)
}

func main() {
	dio.SetBanner("")
	dio.SetProperty("app.env", "dev")
	dio.RegisterBean(&DB{})
	dio.Provide(UserService{})

	// 状态监听：就绪后打印提示
	dio.OnStateChange(func(s dio.AppState) {
		if s == dio.Running {
			fmt.Println("应用已就绪")
		}
	})

	// 优雅停机回调
	dio.OnShutdown(func() {
		fmt.Println("正在清理资源...")
	})

	dio.Run(context.Background())
}
```

## 启动流程

`Run` 内部按以下顺序执行：

1. 进入 `Starting` 状态，打印 banner
2. 校验必填配置项（`RequireProperties`）
3. 创建日志组件并注册容器自身为 bean
4. 按条件装配注册所有 bean，`di.Load()` 实例化并注入
5. 执行 `afterRunFns`，进入 `Running` 状态，打印启动摘要
6. 阻塞等待退出信号 / ctx 取消
7. 销毁 bean（倒序触发 `Destroy`），执行 `OnShutdown` 回调，进入 `Stopped` 状态

## 下一步

- [配置加载](config/loading) — 配置文件、优先级链与 profile
- [状态机](lifecycle/state) — 生命周期状态与就绪探针
- [健康检查](health/health) — 为应用添加健康检查
- [优雅停机](shutdown/shutdown) — 自定义停机行为
