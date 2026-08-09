---
layout: default
title: 概述
nav_order: 1
parent: 入门
---

# 概述

`dio` 是基于 [di](https://github.com/Cheivin/di) 的 **Go 应用层框架**，在底层 DI 容器之上补齐了应用启动所需的通用能力：日志、配置加载、条件装配、信号处理、健康检查与优雅停机。

## 分层架构

```
┌─────────────────────────────────────────────┐
│ 应用代码（你的 Service / Controller）          │
├─────────────────────────────────────────────┤
│ dio（应用层框架）                              │
│   日志 · 配置 · 条件装配 · 状态机 · 健康检查    │
│   优雅停机 · 插件系统                          │
├─────────────────────────────────────────────┤
│ dio-core（接口定义层）                        │
│   Dio / Log / Property / HealthChecker       │
├─────────────────────────────────────────────┤
│ di（底层 DI 容器）                            │
│   依赖注入 · 生命周期 · 批量注入 · 工厂注入      │
└─────────────────────────────────────────────┘
```

- **di**：只负责依赖注入本身，不感知应用概念（不依赖 zap、yaml）
- **dio-core**：定义 dio 与插件共享的接口（`Dio`/`Log`/`Property`），不含实现
- **dio**：`Dio` 接口的实现，包装 di 容器，是应用真正使用的框架
- **插件**：通过 `Use(plugins...)` 注入的适配层（gin/gorm 等），以 `core.PluginConfig` 函数形式注册

## 核心能力

| 能力 | 入口 | 说明 |
|------|------|------|
| 日志 | `SetLogger` / `Logger` | 内置 Zap 实现，自动创建 |
| 配置 | `LoadConfig` / `SetProperty` | YAML 加载 + 优先级链 + profile |
| 条件装配 | `OnProperty` / `OnProfile` / `OnBeanType` | 按配置或已注册类型决定装配 |
| 状态机 | `State` / `OnStateChange` / `Ready` | 应用生命周期六态 |
| 健康检查 | `Health` | 聚合所有 `HealthChecker` bean |
| 优雅停机 | `OnShutdown` | 信号触发，可并行/限时 |
| 插件 | `Use` | 函数式插件装配 |

## 使用方式

dio 支持两种使用方式，见 [全局函数与独立容器](others/global)：

```go
// 方式一：全局函数（推荐，懒初始化）
dio.SetProperty("app.port", 8080).Provide(Service{}).Run(ctx)

// 方式二：独立容器
d := dio.New()
d.SetProperty("app.port", 8080).Provide(Service{}).Run(ctx)
```
