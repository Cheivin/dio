---
layout: default
title: 配置加载
nav_order: 1
parent: 配置
---

# 配置加载

dio 的配置系统基于 di 的 van 存储，支持点号分隔的层级 key（如 `app.port`）。

## 配置来源

### 代码设置

```go
dio.SetDefaultProperty("app.port", 8080)          // 默认配置（低优先级）
dio.SetProperty("app.port", 9090)                 // 显式配置（高优先级）
dio.SetDefaultPropertyMap(map[string]any{...})    // 批量
dio.SetPropertyMap(map[string]any{...})
```

### YAML 文件加载

```go
// 通过 embed 内嵌配置文件
//go:embed configs
var configs embed.FS

dio.LoadDefaultConfig(configs, "configs/default.yaml") // 默认配置
dio.LoadConfig(configs, "configs/config.yaml")         // 公共配置（加载后作为默认级，配合 profile 覆盖）
dio.LoadConfigDir(configs, "configs/dir")              // 加载目录下所有 *.yaml（按文件名排序，后加载覆盖同名项）
```

`LoadConfigDir` 会读取目录下所有 `*.yaml` 文件，按文件名排序依次合并，后加载的文件覆盖同名配置项；子目录与非 yaml 文件会被忽略。

### 环境变量

```go
dio.AutoMigrateEnv() // 读取所有环境变量注入配置，key 中的 _ 转为 .（APP_PORT → app.port）
```

## 优先级链

配置优先级从低到高：

| 优先级 | 来源 | 说明 |
|--------|------|------|
| 低 | `SetDefaultProperty` / `LoadDefaultConfig` | 默认配置 |
| ↑ | `LoadConfig` / `LoadConfigDir` 的公共配置 | 与默认配置同级，加载在后覆盖同名项 |
| ↑ | `LoadConfig` 的 `config-{profile}.yaml` 覆盖配置 | 高于公共配置 |
| ↑ | `AutoMigrateEnv` 环境变量 | 高于配置文件 |
| 高 | `SetProperty` / `SetPropertyMap` 显式配置 | 最高（与 env 同级，后写覆盖先写） |

> 注意：同一级别内"后写覆盖先写"。显式 `SetProperty` 通常在配置链最后调用，因此实际优先级最高。

## 读取配置

```go
// 直接读取
port := dio.GetPropertyString("app.port")

// 判断是否存在（区分"未设置"与"空值"）
if dio.HasProperty("app.port") { ... }

// 映射到结构体（按 value 标签自动转换）
type AppConfig struct {
	Port int    `value:"port"`
	Env  string `value:"env"`
}
cfg := dio.GetProperties("app.", AppConfig{}).(AppConfig)
```

`GetProperties` 支持类型自动转换（int/string/bool 等）。

## 下一步

- [Profile 环境](profile) — 多环境配置覆盖
- [必填校验](require) — 启动时校验必填配置项
