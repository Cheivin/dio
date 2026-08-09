---
layout: default
title: 注册 bean
nav_order: 1
parent: Bean 管理
---

# 注册 bean

dio 继承了 di 的全部注册方式，并在此基础上增加了条件注册。所有注册 API 均可链式调用。

## 注册实例

```go
dio.RegisterBean(&DB{})                 // 按类型推断名称（类型名首字母小写）
dio.RegisterNamedBean("myDB", &DB{})    // 指定名称
dio.RegisterBean(&DB{}, &Cache{})       // 批量
```

## 注册结构体原型

```go
dio.Provide(UserService{})                   // 容器在 Load 时反射实例化为指针并注入依赖
dio.ProvideNamedBean("userService", UserService{})
dio.ProvideMultiNamedBean(map[string]any{    // 按 map 注册多个原型
	"svc1": Service{},
	"svc2": Service{},
})
```

原型注册是推荐方式：容器负责实例化、依赖注入与生命周期回调。

## 注册工厂函数

```go
dio.ProvideFunc(func(db *DB) *UserService {
	return &UserService{DB: db}
})
```

工厂函数的入参按类型自动注入，返回值作为 bean（必须是指针类型）。

## 转发 di 的容器配置

```go
dio.WithCircularCheck(true)      // 开启循环依赖检测（默认关闭，指针循环依赖可正常注入）
dio.WithBeanSelector(selector)   // 设置接口多实现选择策略
```

## 运行期注册（Run 之后）

容器运行后仍可注册，但**只用于按名获取，不经过依赖注入与生命周期回调**：

```go
dio.Run(ctx)
dio.RegisterNamedBean("runtimeBean", obj)
// obj 不会执行 aware/value 注入，也不会触发 BeanConstruct/AfterPropertiesSet 等回调
```

需要完整初始化的 bean 必须在 `Run` 前注册。

## 依赖注入标签

见 di 文档的 [aware](https://cheivin.github.io/di/tag/aware) / [value](https://cheivin.github.io/di/tag/value) 标签说明：

```go
type UserService struct {
	DB  *DB       `aware:""`              // 注入 bean（空名称按类型推断）
	Log core.Log  `aware:""`
	Env string    `value:"app.env"`       // 注入配置项（自动类型转换）
}
```

## 下一步

- [条件装配](condition) — 按配置/类型条件决定是否注册
- [获取与诊断](manage) — 获取 bean 与管理 API
