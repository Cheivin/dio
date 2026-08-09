---
layout: default
title: 获取与诊断
nav_order: 3
parent: Bean 管理
---

# 获取与诊断

## 获取 bean

```go
// 按名获取
svc, ok := dio.GetBean("userService")

// 按类型获取单个（多个时取注册顺序第一个）
bean, ok := dio.GetByType(&UserService{})

// 按类型获取所有（按注册顺序）
all := dio.GetByTypeAll((*Handler)(nil))

// 每次新建实例（非单例，走完整生命周期）
req := dio.NewBean(&RequestContext{}).(*RequestContext)
scope := dio.NewBeanByName("requestScope")
```

类型参数的传法（值类型 / 类型化 nil 指针 / 接口）与 di 完全一致，见 [di 文档](https://cheivin.github.io/di/bean/getbean)。

## Bean 管理 API

除获取实例外，dio 还提供一组只读的管理/诊断 API（转发自 di v0.6.2）：

```go
// 所有 bean 名称（按注册顺序，含工厂 bean）
names := dio.GetBeanNames()

// 是否已注册指定类型的 bean（实例/原型/工厂均可）
if dio.HasBeanType(&UserService{}) { ... }

// bean 定义的只读描述
desc, ok := dio.DescribeBean("userService")
if ok {
	fmt.Println(desc.Name, desc.Type, desc.Factory)
	for _, dep := range desc.Dependencies { // aware 依赖
		fmt.Println(dep.Field, "->", dep.Name)
	}
	for _, v := range desc.Values { // value 配置注入
		fmt.Println(v.Field, "=", v.Name)
	}
}

// 依赖的其他 bean 名称列表（按名称排序）
deps, ok := dio.GetBeanDependencies("userService")
```

`DescribeBean` / `GetBeanDependencies` 的细节与限制见 [di 文档](https://cheivin.github.io/di/bean/getbean)（管理诊断 API 章节）。

## 注意事项

- **单例**：`GetBean` / `GetByType` 多次调用返回同一指针
- **运行期查询**：`GetBean` / `GetByType` / `GetByTypeAll` 基于容器实例，**`Serve` 退出（bean 销毁）后返回空**；`GetBeanNames` / `DescribeBean` / `GetBeanDependencies` 基于定义，不受影响
- **并发安全**：所有查询方法可并发调用
