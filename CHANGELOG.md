# Changelog

本项目遵循 [Semantic Versioning](https://semver.org/lang/zh-CN/)。

## [0.6.2] - 2026-08-09

版本号对齐 di v0.6.2（dio 系列统一对齐 di 的主次版本号）；升级 di v0.6.2 / dio-core v0.6.2。

### 新增

- **应用状态机**：`AppState`（Pending/Starting/Running/Stopping/Stopped/Failed）+ `State()` / `OnStateChange`（回调锁外倒序）；`Ready()` 改为基于状态（== Running）
- **启动信息**：`SetBanner`（传空字符串关闭）/ `StartupDuration()`；进入 Running 状态输出启动摘要（耗时 / bean 数 / profile）
- **配置 Profile**：`SetProfile` / `Profile()`（显式设置优先，其次环境变量 `APP_PROFILE`）；`LoadConfig` 自动加载 `config-{profile}.yaml` 覆盖（公共配置 SetDefault、profile 配置 Set）
- **条件装配增强**：`OnProfile` / `OnBeanType`（立即求值，与 `OnProperty` 风格一致）
- **配置校验**：`RequireProperties` 声明必填配置项，Run 启动时校验，缺失 panic `ErrMissingProperty`（哨兵，可用 errors.Is 判断）
- **配置来源扩展**：`LoadConfigDir`（目录下 *.yaml 按文件名排序合并）+ 配置优先级链文档化（SetDefault < 文件 < 环境变量 < 显式 Set）
- **健康检查**：dio-core 新增 `HealthChecker` 接口；`Health(ctx)` 聚合容器内所有检查器（每检查独立超时、errors.Join 聚合、未就绪返回错误）
- **Bean 管理 API**：`GetBeanNames` / `DescribeBean` / `GetBeanDependencies`（基于 di v0.6.2 新增的只读定义 API）
- **优雅停机增强**：`SetShutdownTimeout` / `SetShutdownParallel`；文档化"OnShutdown 在 bean 销毁后执行"（di.Serve 内部先销毁）

### 修复

- **Run 启动失败可重试的承诺真正成立**：日志创建提前到容器注册之前，日志创建阶段的失败不再污染 di 容器，修正后可重试
- **`ProvideNamedBean` 忽略显式名称**：名称参数被丢弃导致按类型推断，连带 `ProvideMultiNamedBean` 注册多个同类型 bean 时因重名 panic；已修复
- 健康检查聚合跳过容器自身（容器实现 `HealthChecker` 接口，避免自引用无限递归）

### 变更

- `yaml.v2` → `yaml.v3`（嵌套 map 解析为 `map[string]any`，van 兼容）
- 升级 di v0.6.2 / dio-core v0.6.2，版本号对齐

## [0.3.0] - 2026-08-09

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

## [0.2.0]

- 修复日志重复注册问题

## [0.1.1]

- 修复日志重复注册问题

## [0.1.0]

- 初始版本：基于 di 的应用层框架，集成 Zap 日志、YAML 配置、条件装配、插件系统
