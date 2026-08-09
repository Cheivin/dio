---
layout: default
title: 状态机
nav_order: 1
parent: 应用生命周期
---

# 状态机

dio 为应用定义了完整的生命周期状态机（`AppState`），从创建到停止单向推进，可用于就绪探针、指标上报等场景。

## 状态定义

| 状态 | 说明 |
|------|------|
| `Pending` | 容器创建，尚未 `Run` |
| `Starting` | `Run` 开始，正在初始化（日志创建 / bean 注册 / 容器加载） |
| `Running` | 初始化完成，容器就绪，可对外提供服务 |
| `Stopping` | 停机中（`Serve` 退出，正在执行停机回调） |
| `Stopped` | 停机完成 |
| `Failed` | `Run` 启动失败（panic 退出） |

状态只允许**单向推进**：`Pending → Starting → Running → Stopping → Stopped`，失败进入 `Failed`。

## 查询状态

```go
dio.State()          // 当前状态
dio.Ready()          // 是否就绪（等价于 state == Running）
```

`Ready()` 常用于就绪探针：容器未完全启动时不应接收流量。

## 状态变更回调

```go
dio.OnStateChange(func(s dio.AppState) {
	switch s {
	case dio.Running:
		fmt.Println("服务已就绪，开始接收流量")
	case dio.Stopping:
		fmt.Println("开始停机")
	}
})
```

多个回调按**注册倒序**执行（后注册的先执行）；回调在容器锁外调用，可安全访问容器方法。

> 注意：`Starting` 在 bean 注册之前触发，此时容器内容不完整，仅用于记录启动开始事件；需要就绪感知请用 `Running` 状态或 `Ready()`。

## 启动信息

```go
dio.SetBanner("  my app  ")  // 自定义启动横幅（Run 开始打印到控制台）；传空字符串关闭
dio.StartupDuration()        // 启动耗时：Run 开始至今的时间（未 Run 时返回 0）
```

`Run` 进入 `Running` 状态时会输出启动摘要（耗时 / bean 数 / profile）：

```
2026-08-09 16:20:55 INFO  started in 276µs, 2 beans, profile: dev
```

## 启动失败与重试

启动失败（如必填配置缺失、日志创建失败）时状态置为 `Failed`，且 `Run` 还原内部状态：

```go
// 仅"日志创建阶段 / 配置校验阶段"的失败可修正后重试 Run；
// bean 注册 / di.Load 之后的失败，di 容器已残留状态，重试会 panic。
```

`Failed` 后可重新 `Starting`（状态机的唯一回退例外，配合 `Run` 的状态还原实现失败重试）。
