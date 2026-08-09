# dio

`dio` 是基于 [di](https://github.com/Cheivin/di/) 的应用层框架，在 di 容器之上集成了日志、配置加载、条件装配、信号处理和插件系统。

## 特性

- 📦 **基于 di 容器**：继承 di 的全部能力（依赖注入、生命周期、构造函数注入、批量注入等）
- 📝 **内置 Zap 日志**：集成 [zap](https://github.com/uber-go/zap) + [file-rotatelogs](https://github.com/lestrrat-go/file-rotatelogs)，支持日志滚动、文件/控制台分流、请求追踪
- 📄 **YAML 配置加载**：`LoadConfig` / `LoadDefaultConfig` 从 YAML 文件加载配置
- 🔀 **条件装配**：`ProvideOnProperty` / `ProvideNotOnProperty` 按配置项值决定是否注册 bean
- 🔌 **插件系统**：`Use(plugins...)` 函数式插件，gin/gorm 插件可选
- 🔄 **懒初始化**：全局容器首次调用时创建，`Reset()` 可重置（测试隔离）
- ⚡ **信号处理**：`Run` 自动监听 SIGINT/SIGTERM，优雅销毁所有 bean

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

## 全局函数 vs 独立容器

```go
// 全局函数（懒初始化）
dio.Provide(Service{})
dio.Run(ctx)

// 独立容器
d := dio.New()
d.Provide(Service{})
d.Run(ctx)
```

## 配置加载

```go
// embed 内嵌配置文件
//go:embed configs
var configs embed.FS

dio.LoadDefaultConfig(configs, "configs/default.yaml") // 默认配置
dio.LoadConfig(configs, "configs/prod.yaml")           // 覆盖配置
dio.AutoMigrateEnv()                                    // 环境变量（_ 转 .）
```

## 条件装配

```go
dio.ProvideOnProperty(MySQLRepo{}, "db.type", "mysql")   // db.type==mysql 时注册
dio.ProvideNotOnProperty(MockRepo{}, "db.type", "mysql") // db.type!=mysql 时注册
```

## 插件

```go
dio.New().
	SetProperty("app.port", 8080).
	Use(orm.Gorm(), web.GinWeb(true)).
	Provide(MyService{}).
	Run(context.Background())
```

## 相关

- [di](https://github.com/Cheivin/di) — 底层依赖注入容器
- [dio-core](https://github.com/Cheivin/dio-core) — 接口定义层
- [dio-plugin](https://github.com/Cheivin/dio-plugin) — gin/gorm 插件
