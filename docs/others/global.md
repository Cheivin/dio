---
layout: default
title: 全局函数与独立容器
nav_order: 1
parent: 其他
---

# 全局函数与独立容器

dio 提供两种使用方式：包级全局函数（操作一个全局容器）与独立容器实例。

## 全局函数（推荐）

```go
dio.SetProperty("app.port", 8080).
	Provide(Service{}).
	Run(context.Background())
```

- 全局容器**懒初始化**：首次调用任一全局函数时才创建
- 全局函数操作同一个容器实例，模块各处可直接使用
- `dio.Reset()` 可将全局容器重置（**仅用于测试隔离**，生产代码不应调用）：

```go
func TestXxx(t *testing.T) {
	defer dio.Reset() // 测试间隔离，避免状态残留
	...
}
```

## 独立容器

```go
d := dio.New()
d.SetProperty("app.port", 8080)
d.Provide(Service{})
d.Run(ctx)
```

适合需要多个隔离容器或显式管理容器生命周期的场景。

## API 边界（重要）

dio 保持**接口面小**：`core.Dio` 接口只包含基础 API（配置 / 注册 / 获取 / Run 等），**扩展 API 不在接口中**——`OnShutdown` / `Ready` / `State` / `OnStateChange` / `Health` / `SetBanner` / `SetProfile` / `RequireProperties` / `GetBeanNames` / `DescribeBean` / `GetBeanDependencies` 等只有全局函数版本：

```go
dio.OnShutdown(fn)     // ✅ 全局函数
dio.Ready()            // ✅ 全局函数

d := dio.New()
d.OnShutdown(fn)       // ❌ 编译错误：core.Dio 接口无此方法
```

独立容器使用扩展 API 需要类型断言：

```go
d := dio.New().(*dioContainer) // dioContainer 未导出，外部包无法直接断言
```

> 因此：**扩展 API 只对全局容器可用**（通过全局函数）。独立容器场景如需这些能力，请使用全局函数或将容器改为全局使用。这是有意的设计取舍——保持 `core.Dio` 接口稳定（不因新增功能破坏实现方）。

## 运行时注册的注意点

运行期（`Run` 之后）的注册直接进入 di 容器，仅可用于按名获取，**不经过依赖注入与生命周期回调**（见[注册 bean](bean/register)）。
