---
layout: default
title: 错误处理
nav_order: 3
parent: 其他
---

# 错误处理

dio 遵循与 di 一致的错误语义：**运行期错误以 `panic(error)` 形式抛出**，错误值包装哨兵错误，调用方可用 `errors.Is` 判断。

## 错误哨兵

| 哨兵 | 触发场景 |
|------|---------|
| `dio.ErrNotRun` | 容器尚未 `Run` 时调用运行期方法（如 `Logger()`） |
| `dio.ErrAlreadyRun` | 重复 `Run`，或 `Run` 后注册原型/设置日志 |
| `dio.ErrMissingProperty` | `RequireProperties` 声明的必填配置缺失 |

## 捕获与判断

```go
func runApp() (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				panic(r) // 非 error 的 panic 原样抛出
			}
		}
	}()
	dio.Run(ctx)
	return nil
}

if err := runApp(); err != nil {
	switch {
	case errors.Is(err, dio.ErrMissingProperty):
		fmt.Println("配置缺失:", err)
	case errors.Is(err, dio.ErrAlreadyRun):
		fmt.Println("重复启动")
	default:
		fmt.Println("启动失败:", err)
	}
}
```

## 启动失败的可重试性

`Run` 的失败分两种：

- **可重试**：必填配置缺失（`ErrMissingProperty`）、日志创建失败——此时容器未被污染，修正后可直接重新 `Run`（状态机 `Failed → Starting`）
- **不可重试**：bean 注册 / `di.Load` 之后的失败——di 容器已残留状态，重试会 panic（错误为 `ErrAlreadyRun` 或 di 的 `ErrBean`/`ErrDefinition` 等）

## 与 di 的错误

di 的错误哨兵（`ErrBean` / `ErrDefinition` / `ErrLoaded` 等）同样以 `errors.Is` 判断，见 [di 文档](https://cheivin.github.io/di/others/error-handling)。dio 的哨兵独立定义，不重复包装 di 的错误。
