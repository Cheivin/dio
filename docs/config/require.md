---
layout: default
title: 必填校验
nav_order: 3
parent: 配置
---

# 必填校验

`RequireProperties` 声明应用运行必需的配置项，**启动时**（`Run` 中）校验：任一缺失则启动失败。

## 用法

```go
dio.LoadConfig(configs, "configs/config.yaml")
dio.RequireProperties("app.port", "db.host", "db.password")
```

## 校验时机与失败行为

校验发生在 `Run` 的日志创建之前（配置链全部完成之后）。缺失时 `Run` panic：

```go
err := func() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	dio.Run(ctx)
	return nil
}()
if errors.Is(err, dio.ErrMissingProperty) {
	fmt.Println("必填配置缺失:", err)
}
```

panic 的错误包装 `ErrMissingProperty` 哨兵，可用 `errors.Is` 判断（见[错误处理](others/errors)）。

由于校验发生在日志创建与 bean 注册之前，**修正配置后可重新 `Run`**（此时容器未被污染，状态机允许从 `Failed` 恢复）。

## 与条件装配组合

profile 等条件装配可以先注册，再统一校验：

```go
dio.SetProfile("prod").
	RequireProperties("app.port", "db.host").
	Run(ctx)
```
