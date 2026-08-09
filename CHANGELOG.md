# Changelog

本项目遵循 [Semantic Versioning](https://semver.org/lang/zh-CN/)。

## [0.2.0] - 2026-08-09

### Breaking Changes

- **升级 di 至 v0.6.1**：同步 di 的全部 breaking 变更（Log.Fatal(error)、Go 1.25、DI 接口新方法等）
- **升级 dio-core 至 v0.2.0**：`Dio` 接口新增 `WithCircularCheck(enable bool) Dio`
- **新增方法**：`dioContainer` 实现 `WithCircularCheck`、`WithBeanSelector`、`ProvideFunc`（对齐 di 新能力）

### 变更

- `interface{}` 全量迁移为 `any`
- 修复 `dio_logger.go` 的 Fatal 实现（不再 os.Exit，改为 panic + error）
- 全局容器懒初始化（首次调用才创建），新增 `dio.Reset()` 测试隔离
- 修复测试阻塞问题：`Run(context.Background())` 改为可取消 ctx
- README 补充完整文档

## [0.1.1]

- 修复日志重复注册问题

## [0.1.0]

- 初始版本：基于 di 的应用层框架，集成 Zap 日志、YAML 配置、条件装配、插件系统
