---
layout: default
title: 健康检查
nav_order: 1
parent: 健康检查
---

# 健康检查

dio 通过 `core.HealthChecker` 接口 + `Health()` 聚合方法为应用提供统一健康检查能力，适合对接 k8s 探针 / 网关健康检查。

## 实现 HealthChecker

任何 bean 实现 `Health(ctx) error` 即成为健康检查器，返回 `nil` 表示健康：

```go
import "github.com/cheivin/dio-core"

type DatabaseHealth struct {
	DB *DB `aware:""`
}

func (h *DatabaseHealth) Health(ctx context.Context) error {
	// 检查数据库连通性；超时通过 ctx 感知
	return h.DB.Ping(ctx)
}
```

```go
dio.Provide(DatabaseHealth{})
```

## 聚合检查

```go
err := dio.Health(ctx)
if err != nil {
	// 不健康：err 聚合了所有失败检查器的错误（errors.Join，可用 errors.Is 判断）
}
```

`Health()` 的行为：

- **未就绪门控**：容器不在 `Running` 状态（未启动 / 停机中）时直接返回错误，避免误报健康
- **聚合**：遍历容器内所有实现 `HealthChecker` 的 bean，逐个执行，收集全部失败（`errors.Join`）
- **每检查超时**：单个检查有独立超时（5 秒）；传入的 `ctx` 若带 deadline 会进一步收紧
- **错误信息**：每个失败带 bean 名称，如 `health check failed for databaseHealth: ...`

## 典型用法（HTTP 探针）

配合 gin 插件暴露 `/health`：

```go
func HealthHandler(c *gin.Context) {
	if err := dio.Health(c.Request.Context()); err != nil {
		c.JSON(503, gin.H{"status": "DOWN", "detail": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "UP"})
}
```

> 提示：`Ready()`（状态机）与 `Health()`（业务检查）是互补的——前者反映容器启动状态，后者反映业务依赖可用性。就绪探针通常两者都看。
