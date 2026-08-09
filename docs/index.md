# dio

`dio` 是基于 [di](https://github.com/Cheivin/di) 的应用层框架，在 di 容器之上集成了日志、配置加载、条件装配、信号处理、健康检查、状态机与插件系统。

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## 特性

- 📦 **基于 di 容器**：继承 di 的全部能力（依赖注入、生命周期、构造函数注入、批量注入等）
- 📝 **内置 Zap 日志**：集成 [zap](https://github.com/uber-go/zap) + [file-rotatelogs](https://github.com/lestrrat-go/file-rotatelogs)，支持日志滚动、文件/控制台分流、请求追踪
- 📄 **YAML 配置加载**：`LoadConfig` / `LoadDefaultConfig` / `LoadConfigDir`，支持 profile 覆盖与优先级链
- 🗂️ **Profile 环境**：`SetProfile` / `APP_PROFILE` 环境变量，自动加载 `config-{profile}.yaml` 覆盖配置
- ✅ **配置校验**：`RequireProperties` 启动时校验必填配置项，缺失即失败
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

## 文档目录

### 入门

- [概述](quickstart/introduction) — dio 是什么、分层架构
- [安装](quickstart/install) — 安装与版本要求
- [快速开始](quickstart/quickstart) — 5 分钟上手

### 配置

- [配置加载](config/loading) — 加载方式与优先级链
- [Profile 环境](config/profile) — 多环境配置
- [必填校验](config/require) — RequireProperties

### Bean 管理

- [注册 bean](bean/register) — RegisterBean / Provide / ProvideFunc
- [条件装配](bean/condition) — OnProperty / OnProfile / OnBeanType
- [获取与诊断](bean/manage) — GetBean / GetByType / Bean 管理 API

### 应用生命周期

- [状态机](lifecycle/state) — AppState / State / OnStateChange / Ready

### 健康检查

- [健康检查](health/health) — HealthChecker / Health

### 优雅停机

- [优雅停机](shutdown/shutdown) — OnShutdown / 超时 / 并行

### 其他

- [全局函数与独立容器](others/global) — API 边界与 Reset
- [日志配置](others/log) — Zap 日志与请求追踪
- [错误处理](others/errors) — 错误哨兵与 errors.Is

### 参考

- [更新日志](https://github.com/cheivin/dio/blob/main/CHANGELOG.md)
- [di 文档](https://cheivin.github.io/di/) — 底层依赖注入容器
