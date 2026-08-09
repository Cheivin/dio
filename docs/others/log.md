---
layout: default
title: 日志配置
nav_order: 2
parent: 其他
---

# 日志配置

dio 内置基于 [zap](https://github.com/uber-go/zap) + [file-rotatelogs](https://github.com/lestrrat-go/file-rotatelogs) 的日志组件，支持日志滚动、文件/控制台分流与请求追踪。

## 默认配置

`Run` 时若未通过 `SetLogger` 指定日志组件，dio 会按 `log` 前缀配置自动创建：

```go
dio.SetDefaultProperty("log", map[string]any{
	"name":       "dio_app",  // 日志文件名（含 @hostname 会替换为主机名）
	"dir":        "./logs",   // 日志目录
	"max-age":    30,         // 日志保留天数
	"debug":      true,       // 是否输出 DEBUG 级别（false 时从 INFO 起）
	"std":        true,       // 输出到控制台
	"file":       true,       // 输出到文件
	"trace-name": "X-Request-Id", // 请求追踪 key
})
```

文件输出会生成 `{name}.log`（INFO 及以上）与 `{name}_error.log`（ERROR 及以上）两个滚动文件。`file` 与 `std` 都关闭时会强制开启控制台输出。

## 使用日志

bean 通过 `aware` 标签注入日志：

```go
type Service struct {
	Log core.Log `aware:""`
}

func (s *Service) Do(ctx context.Context) {
	s.Log.Info(ctx, "开始处理", "orderId", 1001)      // 结构化字段
	s.Log.Errorw(ctx, "处理失败", map[string]any{     // map 形式字段
		"orderId": 1001,
	})
	s.Log.Debug(ctx, "调试信息")
}
```

`core.Log` 接口方法见 [dio-core](https://github.com/Cheivin/dio-core)（`Debug/Info/Warn/Error` 及其 `w` 变体）。

## 请求追踪

```go
// 在请求入口生成/注入 traceId，后续日志自动带上
ctx = s.Log.Trace(ctx)              // 生成新 traceId 并写入 ctx
ctx = s.Log.TraceWith(ctx, reqId)   // 使用指定 traceId

// 之后所有带 ctx 的日志都会输出 traceId 字段
s.Log.Info(ctx, "处理请求")
```

traceId 的 key 由 `log.trace-name` 配置决定（默认 `X-Request-Id`），可与中间件透传的请求头对齐。

## 自定义日志组件

```go
// 使用其他日志实现（实现 core.Log 接口）
dio.SetLogger(myLogger)

// 或包装现有 zap logger
zapLogger := zap.NewExample()
dio.SetLogger(dio.WrapZapLogger(zapLogger))
```

```go
// 获取当前日志组件
log := dio.Logger() // 必须在 Run 之后（Run 前会 panic：ErrNotRun）
```

## 日志组件生命周期

- `Run` 启动失败时，已创建的日志组件会被自动关闭（避免文件句柄泄漏）
- 停机时（bean 销毁阶段）日志组件 flush 所有缓冲后关闭——因此停机回调里的日志仍能正常输出
