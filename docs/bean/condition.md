---
layout: default
title: 条件装配
nav_order: 2
parent: Bean 管理
---

# 条件装配

dio 提供三类条件装配：按配置项（`OnProperty`）、按 profile（`OnProfile`）、按已注册类型（`OnBeanType`）。

## 立即执行（按条件执行代码）

以下方法**立即求值**：条件满足时当场执行 `fn`，不满足则跳过。适合"根据条件再注册一批 bean"的场景：

```go
// 按配置项
dio.OnProperty("db.type", "mysql", false, func(d core.Dio) {
	d.Provide(MySQLRepo{})
})
dio.NotOnProperty("db.type", "mysql", false, func(d core.Dio) {
	d.Provide(MockRepo{})
})

// 按 profile（解析规则见 Profile 文档）
dio.OnProfile("dev", func(d core.Dio) {
	d.Provide(DebugMiddleware{})
})

// 按已注册类型（实例/原型/工厂 bean 均可命中，接口类型也支持）
dio.OnBeanType((*Notifier)(nil), func(d core.Dio) {
	d.Provide(SmsFallback{})
})
```

`OnProperty` 的 `caseSensitive` 参数控制比较值是否大小写敏感。

## 条件注册（注册时携带条件）

以下方法把条件记录在 bean 定义上，**`Run` 时统一判断**：

```go
// 配置项条件
dio.ProvideOnProperty(MySQLRepo{}, "db.type", "mysql")
dio.ProvideNamedBeanOnProperty("repo", MySQLRepo{}, "db.type", "mysql")
dio.ProvideNotOnProperty(MockRepo{}, "db.type", "mysql")

// 批量
dio.ProvideMultiBeanOnProperty([]any{RepoA{}, RepoB{}}, "db.type", "mysql")
dio.ProvideMultiNamedBeanOnProperty(map[string]any{"a": RepoA{}}, "db.type", "mysql")
```

条件不满足的 bean 不会进入容器。

## 选择哪种？

| 场景 | 用哪个 |
|------|--------|
| 条件满足时注册一组 bean | `OnProperty` / `OnProfile` / `OnBeanType`（立即执行） |
| 单个 bean 带条件注册 | `ProvideOnProperty` 系列（Run 时统一判断） |
| 在 `Run` 之后判断配置 | 都不行——条件装配是启动期机制，运行期用代码自行判断 |
