---
layout: default
title: 升级指南
nav_order: 4
parent: 其他
---

# 升级指南

本文档说明从旧版本升级到 dio v0.6.2（dio 系列 0.6.2 全家桶）时的破坏性变更与迁移动作。

## 版本配套关系

dio 系列统一对齐 di 的主次版本号，升级时各模块应配套升级：

| 模块 | 版本 |
|------|------|
| di | v0.6.2 |
| dio-core | v0.6.2 |
| dio | v0.6.2 |
| dio-plugin（gin/gorm） | v0.6.2（发布后） |

## 破坏性变更总览

> **结论**：破坏面全部来自历史版本（di v0.4.0/v0.5.0 与 dio v0.3.0 时代的接口演进），**dio v0.6.2 与 dio-core v0.6.2 本身零破坏**（核心接口未动，9 个 Feature 纯新增）。

按影响程度排序：

| 严重度 | 变更 | 适用版本升级路径 |
|--------|------|-----------------|
| 高 | Go 最低版本 1.25 | 所有路径 |
| 高 | `Log.Fatal` 签名 `string` → `error` | di <0.4.0、dio <0.3.0 |
| 高 | `Fatal` 行为从 `os.Exit` 改为 panic | 所有路径 |
| 中 | `LoadConfig` 公共配置优先级调整 | dio <0.6.2 |
| 中 | `yaml.v2` → `yaml.v3` | dio <0.6.2 |
| 中 | 循环依赖检测默认关闭 | di 0.3.1~0.5.0 |
| 低 | `GetByTypeAll` 返回顺序修正 | di <0.6.2 |
| 低 | van merge/cast 行为修正 | di <0.4.0 |
| 低 | 全局容器懒初始化 / `Reset()` | di <0.4.0、dio <0.3.0 |

## 逐项说明与迁移动作

### 1. Go 1.25+（高）

di v0.4.0 起使用 `slices`/`maps` 标准库特性，最低版本提升至 **Go 1.25**。

**动作**：升级构建环境与 CI 工具链。

### 2. `Log.Fatal` 签名变更（高）

`Fatal(string)` → `Fatal(error)`（di v0.4.0，dio v0.3.0 同步）。

```go
// 旧
func (l *MyLog) Fatal(s string)

// 新
func (l *MyLog) Fatal(err error)
```

**动作**：自定义 `core.Log` / `di.Log` 实现同步修改签名。`recover()` 后可对 panic 值使用 `errors.Is` 判断具体错误（`ErrBean`/`ErrDefinition`/`ErrLoaded` 等）。

### 3. `Fatal` 行为：不再 os.Exit（高）

di v0.4.0 起，标准 logger 与 dio logger 的 `Fatal` 从 `os.Exit` 改为 **panic**：

- 错误可被 `recover` 捕获（不会直接杀死进程）
- 未 recover 时输出堆栈而非静默退出

**动作**：检查依赖"Fatal 即退出"的部署逻辑（如 supervisor 依赖退出码判断）。

### 4. `LoadConfig` 公共配置优先级调整（中，dio v0.6.2）

`LoadConfig` 的公共配置从 Set 级（高优先级）调整为 SetDefault 级（低优先级），profile 覆盖配置（`config-{profile}.yaml`）使用 Set 级。优先级链见[配置加载](config/loading)。

**⚠️ 特定调用顺序行为变化**：

```go
dio.SetProperty("app.port", 9090)   // 先显式设置
dio.LoadConfig(configs, "config.yaml") // 后加载文件（app.port: 8080）
// 旧版本：文件覆盖代码 → app.port = 8080
// v0.6.2：代码覆盖文件 → app.port = 9090
```

常见调用顺序（先 `LoadConfig` 后 `SetProperty`）行为不变。

**动作**：检查代码中 `SetProperty` 之后调用 `LoadConfig` 的覆盖预期。

### 5. `yaml.v2` → `yaml.v3`（中，dio v0.6.2）

嵌套 map 的解析结果从 `map[interface{}]interface{}` 变为 `map[string]any`（van 兼容，配置读取无感）。

**动作**：若对 yaml 解析结果直接做 `map[interface{}]interface{}` 类型断言（例如从 `LoadConfig` 拿原始 map 的场景），需改为 `map[string]any`。常规 `GetPropertyString` / `GetProperties` 用法不受影响。

### 6. 循环依赖检测默认关闭（中，di 0.3.1~0.5.0 → 0.6.x）

v0.3.1 至 v0.5.0 曾默认开启循环依赖检测（会拦截本可正常工作的指针循环依赖，属回归 bug，v0.4.0/v0.5.0 已 retract）。v0.6.1 起**默认关闭**：

- 指针循环依赖（如 `A.B.A == A`）恢复正常注入
- 需要保证依赖关系为 DAG 的严格场景，显式开启：`dio.WithCircularCheck(true)`

### 7. `GetByTypeAll` 返回顺序修正（低，di v0.6.2）

此前遍历 map 取实例，返回顺序随机（与接口注释不符）；现改为**按注册顺序**返回，`GetByType` 返回第一个注册的匹配 bean。

**动作**：依赖"注册顺序"的代码现在行为符合预期；依赖随机序的代码（理论上不应存在）需调整。

### 8. van merge/cast 行为修正（低，di v0.4.0）

- `mergeStringMap` 类型冲突（string vs map）时：新值覆盖旧值（原静默丢弃）
- cast 未知类型：`fmt.Sprint` 兜底（不再静默丢值）

**动作**：边缘场景，检查配置合并结果是否符合预期。

### 9. 全局容器懒初始化 / `Reset()`（低）

`import` 不再创建容器，首次调用全局函数时才创建（di v0.4.0、dio v0.3.0）；新增 `Reset()` 供测试隔离。

**动作**：无（正常代码无感知）；测试代码可用 `defer dio.Reset()` 隔离状态。

## 接口实现方特别提示

`DI` / `Dio` 接口历次新增方法（`ProvideFunc`/`WithBeanSelector`/`AutoMigrateEnv`/`WithCircularCheck`/`GetByTypeAll`/`Context`/`GetBeanNames`/`HasBeanType`/`DescribeBean`/`GetBeanDependencies` 等），外部实现这些接口的类型需要补充实现。生态中仅框架自身实现，第三方实现罕见；若有则按编译错误逐个补齐（均可在 di 文档找到签名）。

## 升级检查清单

- [ ] Go 工具链 ≥ 1.25
- [ ] 自定义 `Log` 实现：`Fatal(error)` 签名 + panic 语义
- [ ] 依赖退出码判断 `Fatal` 的部署逻辑
- [ ] `SetProperty` 后 `LoadConfig` 的覆盖预期（v0.6.2 变化）
- [ ] yaml 原始 map 类型断言（`map[string]any`）
- [ ] 循环依赖检测需求（`WithCircularCheck(true)`）
- [ ] 配套升级 di / dio-core / dio / 插件到 0.6.2
