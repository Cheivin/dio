---
layout: default
title: 优雅停机
nav_order: 1
parent: 优雅停机
---

# 优雅停机

`Run` 阻塞等待退出信号（SIGINT / SIGTERM）或 ctx 取消，退出时按固定顺序执行停机流程。

## 停机流程

```
收到信号 / ctx 取消
  └─ di.Serve 退出
      └─ 倒序销毁 bean（触发各 bean 的 Destroy 回调）
          └─ 进入 Stopping 状态
              └─ 执行 OnShutdown 回调（默认倒序，可并行/限时）
                  └─ 进入 Stopped 状态
```

> ⚠️ **顺序很重要**：`OnShutdown` 回调在 **bean 销毁之后**执行（di.Serve 内部先销毁 bean）。因此回调内不应再访问容器读 API（`GetBean` 等已返回空），只做资源清理（关连接池、刷新缓冲）。

## 注册停机回调

```go
dio.OnShutdown(func() {
	dbPool.Close()
	redis.Close()
})
```

多个回调按**注册倒序**执行（后注册的先执行）——与 bean 销毁的倒序语义一致。

## 超时控制

```go
dio.SetShutdownTimeout(10 * time.Second) // 停机回调总超时，0 表示不限时（默认）
```

超时语义（按执行模式）：

| 模式 | 行为 |
|------|------|
| 顺序（默认） | 每个回调执行前检查时限，超时则**跳过剩余回调** |
| 并行 | 等待全部完成或超时；超时后不再等待，未完成回调在后台继续执行 |

> 限制：单个回调自身阻塞无法被中断（Go 无法强制终止 goroutine）；超时仅约束 `OnShutdown` 回调阶段，bean 销毁（di.Serve 内部）不受此限制。

## 并行执行

```go
dio.SetShutdownParallel(true) // 所有回调并发执行并等待完成
```

适合多个互相独立的资源清理（每个耗时较长时，并行能显著缩短停机时间）。

> 并行 + 超时后，未完成的回调继续在后台执行——此时容器 bean 已销毁，回调内不应访问容器读 API。

## 完整示例

```go
dio.SetShutdownTimeout(5 * time.Second).
	SetShutdownParallel(true).
	OnShutdown(func() { dbPool.Close() }).
	OnShutdown(func() { cache.Flush() }).
	Run(ctx)
```
