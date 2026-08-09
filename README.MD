# dio

`dio` 是基于 [di](https://github.com/Cheivin/di/) 的应用层框架，在 di 容器之上集成了日志、配置加载、条件装配、信号处理、健康检查、状态机与插件系统。

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

> 📖 完整文档见 [dio docs](https://cheivin.github.io/dio/)

## 特性

- 📦 **基于 di 容器**：继承 di 的全部能力（依赖注入、生命周期、构造函数注入、批量注入等）
- 📝 **内置 Zap 日志**：集成 [zap](https://github.com/uber-go/zap) + [file-rotatelogs](https://github.com/lestrrat-go/file-rotatelogs)，支持日志滚动、文件/控制台分流、请求追踪
- 📄 **YAML 配置加载**：`LoadConfig` / `LoadDefaultConfig` / `LoadConfigDir`，支持 profile 覆盖与优先级链
- 🗂️ **Profile 环境**：`SetProfile` / `APP_PROFILE`，自动加载 `config-{profile}.yaml` 覆盖配置
- ✅ **配置校验**：`RequireProperties` 启动时校验必填配置项
- 🔀 **条件装配**：`OnProperty` / `OnProfile` / `OnBeanType` 按条件注册 bean
- ⚙️ **应用状态机**：`AppState` 六态生命周期，`State()` / `OnStateChange` / `Ready()`
- 💓 **健康检查**：`HealthChecker` 接口 + `Health()` 聚合检查
- 🛑 **优雅停机**：`OnShutdown` 回调（可并行/限时），bean 倒序销毁
- 🚀 **启动信息**：`SetBanner` 自定义启动横幅，`StartupDuration()` 启动耗时
- 🔌 **插件系统**：`Use(plugins...)` 函数式插件，gin/gorm 插件可选
- 🔄 **懒初始化**：全局容器首次调用时创建，`Reset()` 可重置（测试隔离）

## 快速开始

```go
package main

import (
	"context"

	"github.com/cheivin/dio"
)

type Service struct {
	// 依赖通过 aware 标签注入
}

func main() {
	dio.SetProperty("app.port", 8080).
		Provide(Service{}).
		Run(context.Background())
}
```

`Run` 阻塞等待退出信号（SIGINT/SIGTERM）或 ctx 取消，退出时自动销毁 bean 并执行停机回调。

## 全局函数 vs 独立容器

```go
// 全局函数（懒初始化，推荐）
dio.Provide(Service{})
dio.Run(ctx)

// 独立容器
d := dio.New()
d.Provide(Service{})
d.Run(ctx)
```

> 注意：扩展 API（`OnShutdown` / `Ready` / `State` / `Health` 等）仅提供全局函数版本，不在 `core.Dio` 接口中——见 [API 边界说明](https://cheivin.github.io/dio/others/global)。

## 配置加载

```go
// embed 内嵌配置文件
//go:embed configs
var configs embed.FS

dio.LoadDefaultConfig(configs, "configs/default.yaml") // 默认配置
dio.LoadConfig(configs, "configs/config.yaml")         // 公共配置（配合 profile 覆盖）
dio.LoadConfigDir(configs, "configs/dir")              // 目录下所有 *.yaml 合并
dio.AutoMigrateEnv()                                    // 环境变量（_ 转 .）

dio.SetProfile("dev")                                   // 自动加载 config-dev.yaml 覆盖
dio.RequireProperties("app.port")                       // 必填配置校验
```

## 条件装配

```go
dio.ProvideOnProperty(MySQLRepo{}, "db.type", "mysql")   // db.type==mysql 时注册
dio.ProvideNotOnProperty(MockRepo{}, "db.type", "mysql") // db.type!=mysql 时注册
dio.OnProfile("dev", func(d core.Dio) { ... })           // profile 条件执行
dio.OnBeanType((*Notifier)(nil), func(d core.Dio) { ... }) // 按已注册类型条件执行
```

## 插件

```go
dio.New().
	SetProperty("app.port", 8080).
	Use(orm.Gorm(), web.GinWeb(true)).
	Provide(MyService{}).
	Run(context.Background())
```

## 文档

- [dio docs](https://cheivin.github.io/dio/) — 完整文档（配置 / Profile / 条件装配 / 状态机 / 健康检查 / 优雅停机）
- [di](https://github.com/Cheivin/di) — 底层依赖注入容器
- [dio-core](https://github.com/Cheivin/dio-core) — 接口定义层
- [dio-plugin](https://github.com/Cheivin/dio-plugin) — gin/gorm 插件
