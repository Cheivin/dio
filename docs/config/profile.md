---
layout: default
title: Profile 环境
nav_order: 2
parent: 配置
---

# Profile 环境

Profile（环境）用于区分不同的运行环境（dev / test / prod 等），让同一个应用在不同环境下加载不同的配置。

## 设置 Profile

```go
dio.SetProfile("dev") // 显式设置
```

也可以不写代码，通过环境变量指定：

```bash
APP_PROFILE=prod ./app
```

解析优先级：**显式 `SetProfile` > 环境变量 `APP_PROFILE` > 空**。

```go
dio.Profile() // 返回当前生效的 profile
```

## 自动加载 profile 覆盖配置

当设置了 profile 时，`LoadConfig` 会自动尝试加载 `config-{profile}.yaml` 作为**高优先级覆盖**：

```yaml
# config.yaml（公共配置，低优先级）
app:
  env: prod
  port: 8080

# config-dev.yaml（profile 覆盖，高优先级）
app:
  env: dev
```

```go
dio.SetProfile("dev")
dio.LoadConfig(configs, "configs/config.yaml")
// app.env = dev（来自 config-dev.yaml 覆盖）
// app.port = 8080（来自 config.yaml 公共配置）
```

profile 覆盖文件**不存在时静默忽略**，不报错。文件命名规则：`config.yaml` + profile=`dev` → `config-dev.yaml`。

> 注意：`LoadConfigDir` 加载目录配置时不区分 profile；需要按 profile 覆盖请使用 `LoadConfig` 的 `config-{profile}.yaml` 约定。

## 条件装配

profile 也是[条件装配](bean/condition)的输入之一：

```go
dio.OnProfile("dev", func(d core.Dio) {
	d.Provide(DevRepository{})
})
```
