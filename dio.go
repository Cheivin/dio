package dio

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/signal"
	"path"
	"reflect"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/cheivin/di"
	"github.com/cheivin/dio-core"
	"gopkg.in/yaml.v3"
)

type dioContainer struct {
	log            core.Log
	di             di.DI
	providedBeans  []bean
	loaded         bool
	shutdownFns    []func()         // 优雅停机回调（Serve 退出后倒序执行）
	state          AppState         // 应用生命周期状态（Pending→…→Stopped/Failed 单向推进）
	stateChangeFns []func(AppState) // 状态变更回调（状态推进时快照后锁外倒序执行）
	banner         string           // 启动 banner（空串不打印）
	startTime      time.Time        // Run 开始时间（启动耗时统计起点）
	profile        string           // 显式设置的 profile（优先于环境变量 APP_PROFILE）
	requiredProps  []string         // 必填配置项（RequireProperties 声明，Run 启动时校验）
	shutdownTimeout time.Duration   // 停机回调（OnShutdown）总超时，0 表示不限时
	shutdownParallel bool           // 停机回调是否并行执行（默认顺序倒序）
	mu             sync.Mutex       // 保护 providedBeans/shutdownFns/state/stateChangeFns/startTime/requiredProps 的并发读写
}

// AppState 应用生命周期状态，随容器启动/停机单向推进。
type AppState int

// 应用生命周期状态：
//   - Pending：容器创建，尚未 Run
//   - Starting：Run 开始，正在初始化（日志创建/bean 注册/容器加载）
//   - Running：初始化完成，容器就绪，可对外提供服务
//   - Stopping：停机中（Serve 退出，正在执行停机回调）
//   - Stopped：停机完成
//   - Failed：Run 启动失败（panic 退出）
const (
	Pending AppState = iota
	Starting
	Running
	Stopping
	Stopped
	Failed
)

func (s AppState) String() string {
	switch s {
	case Pending:
		return "Pending"
	case Starting:
		return "Starting"
	case Running:
		return "Running"
	case Stopping:
		return "Stopping"
	case Stopped:
		return "Stopped"
	case Failed:
		return "Failed"
	default:
		return fmt.Sprintf("AppState(%d)", int(s))
	}
}

type bean struct {
	name            string      // 名称
	instance        any // 实例
	needMatch       bool        // 是否条件载入
	property        string      // 条件载入配置项
	compareValue    string      // 条件载入配置比较值
	caseInsensitive bool        // 条件载入配置比较值大小写敏感
	registered      bool        // 是否为手动注册的bean
}

// dio 层错误哨兵。
// 与 di 的 panic(error) 语义保持一致，调用方可用 errors.Is 判断。
var (
	// ErrNotRun 容器尚未 Run（调用 Logger 等运行期方法时）
	ErrNotRun = errors.New("dio not run")
	// ErrAlreadyRun 容器已 Run（重复 Run 或 Run 后注册原型时）
	ErrAlreadyRun = errors.New("dio already run")
	// ErrMissingProperty 必填配置项缺失（RequireProperties 校验未通过）
	ErrMissingProperty = errors.New("dio missing property")
)

// 默认启动 banner，可通过 SetBanner 自定义或传空字符串关闭
const defaultBanner = " ____    ______   _____      \n/\\  _`\\ /\\__  _\\ /\\  __`\\    \n\\ \\ \\/\\ \\/_/\\ \\/ \\ \\ \\/\\ \\   \n \\ \\ \\ \\ \\ \\ \\ \\  \\ \\ \\ \\ \\  \n  \\ \\_\\ \\ \\_\\ \\__\\ \\ \\_\\ \\ \n   \\ \\____/ /\\_____\\\\ \\_____\\\n    \\/___/  \\/_____/ \\/_____/"

// healthCheckTimeout 单个健康检查的超时时间。
// 可通过传给 Health 的 ctx deadline 收紧（整体时限），但不能放宽。
const healthCheckTimeout = 5 * time.Second

func New() core.Dio {
	container := &dioContainer{di: di.New(), providedBeans: []bean{}, loaded: false, banner: defaultBanner}
	container.di.Log(emptyLogger{})
	logName := "dio_app"
	if hostname, err := os.Hostname(); err == nil && hostname != "" {
		logName += "_" + hostname
	}
	container.SetDefaultProperty("log", map[string]any{
		"name":       logName,
		"dir":        "./logs",
		"max-age":    30,
		"debug":      true,
		"std":        true,
		"file":       true,
		"trace-name": core.DefaultTraceName,
	})
	return container
}

func (b bean) matchProperty(d *dioContainer) (match bool) {
	return d.matchProperty(b.property, b.compareValue, b.needMatch, b.caseInsensitive)
}

func (d *dioContainer) matchProperty(property string, compareValue string, needMatch bool, caseInsensitive bool) (match bool) {
	// 空值表示未设定条件
	if property == "" {
		return true
	}
	// 取出比较的属性值
	val := d.di.Property().Get(property)
	if val == nil {
		match = compareValue == ""
	} else {
		propertyValue := fmt.Sprintf("%v", val)
		if caseInsensitive {
			match = strings.EqualFold(propertyValue, compareValue)
		} else {
			match = propertyValue == compareValue
		}
	}
	// needMatch=true 时取 match（On 语义），false 时取反（NotOn 语义）
	return match == needMatch
}

func (d *dioContainer) SetDefaultProperty(key string, value any) core.Dio {
	d.di.SetDefaultProperty(key, value)
	return d
}

func (d *dioContainer) SetDefaultPropertyMap(properties map[string]any) core.Dio {
	d.di.SetDefaultPropertyMap(properties)
	return d
}

func (d *dioContainer) SetProperty(key string, value any) core.Dio {
	d.di.SetProperty(key, value)
	return d
}

func (d *dioContainer) SetPropertyMap(properties map[string]any) core.Dio {
	d.di.SetPropertyMap(properties)
	return d
}

func (d *dioContainer) HasProperty(property string) bool {
	return d.di.Property().Get(property) != nil
}

func (d *dioContainer) GetPropertyString(property string) string {
	val := d.di.Property().Get(property)
	if val == nil {
		return ""
	} else {
		return fmt.Sprintf("%v", val)
	}
}

func (d *dioContainer) GetProperties(prefix string, destType any) any {
	return d.di.LoadProperties(prefix, destType)
}

func (d *dioContainer) AutoMigrateEnv() core.Dio {
	d.di.AutoMigrateEnv()
	return d
}

func (d *dioContainer) SetLogger(log core.Log) core.Dio {
	if d.loaded {
		panic(fmt.Errorf("%w: dioContainer is already run", ErrAlreadyRun))
	}
	d.log = log
	d.di.RegisterBean(log)
	return d
}

func (d *dioContainer) Logger() core.Log {
	if !d.loaded {
		panic(fmt.Errorf("%w: dioContainer not run", ErrNotRun))
	}
	return d.log
}

func (d *dioContainer) RegisterBean(beanInstance ...any) core.Dio {
	for _, bean := range beanInstance {
		d.RegisterNamedBean("", bean)
	}
	return d
}

func (d *dioContainer) RegisterNamedBean(beanName string, beanInstance any) core.Dio {
	if d.loaded {
		// 容器已运行：直接注册到 di（仅可用于按名获取，不经过依赖注入与生命周期回调）。
		// 注意：运行期注册的实例不会执行 aware/value 注入和 BeanConstruct/AfterPropertiesSet 等回调，
		// 需要完整初始化的 bean 应在 Run 前注册。
		d.di.RegisterNamedBean(beanName, beanInstance)
	} else {
		d.mu.Lock()
		d.providedBeans = append(d.providedBeans,
			bean{name: beanName,
				instance:   beanInstance,
				needMatch:  false,
				registered: true,
			})
		d.mu.Unlock()
	}
	return d
}

func (d *dioContainer) Provide(prototype ...any) core.Dio {
	for _, bean := range prototype {
		d.ProvideNamedBean("", bean)
	}
	return d
}

func (d *dioContainer) ProvideNamedBean(beanName string, prototype any) core.Dio {
	return d.ProvideNamedBeanOnProperty(beanName, prototype, "", "")
}

func (d *dioContainer) ProvideFunc(fn any) core.Dio {
	if d.loaded {
		panic(fmt.Errorf("%w: dioContainer is already run", ErrAlreadyRun))
	}
	d.di.ProvideFunc(fn)
	return d
}

// WithCircularCheck 开启/关闭循环依赖检测，转发到底层 di 容器
func (d *dioContainer) WithCircularCheck(enable bool) core.Dio {
	d.di.WithCircularCheck(enable)
	return d
}

// WithBeanSelector 设置接口多实现选择策略，转发到底层 di 容器
func (d *dioContainer) WithBeanSelector(s di.BeanSelector) core.Dio {
	d.di.WithBeanSelector(s)
	return d
}

func (d *dioContainer) ProvideMultiNamedBean(namedBeanMap map[string]any) core.Dio {
	for name, bean := range namedBeanMap {
		d.ProvideNamedBean(name, bean)
	}
	return d
}

func (d *dioContainer) provideBeanCondition(beanName string, prototype any, property string, compareValue string, needMatch bool, caseSensitive bool) core.Dio {
	if d.loaded {
		panic(fmt.Errorf("%w: dioContainer is already run", ErrAlreadyRun))
	}
	d.mu.Lock()
	d.providedBeans = append(d.providedBeans,
		bean{name: beanName,
			instance:        prototype,
			property:        property,
			compareValue:    compareValue,
			needMatch:       needMatch,
			caseInsensitive: !caseSensitive,
		})
	d.mu.Unlock()
	return d
}

func (d *dioContainer) OnProperty(property string, compareValue string, caseSensitive bool, fn func(core.Dio)) core.Dio {
	if d.matchProperty(property, compareValue, true, !caseSensitive) {
		fn(d)
	}
	return d
}

func (d *dioContainer) NotOnProperty(property string, compareValue string, caseSensitive bool, fn func(core.Dio)) core.Dio {
	if d.matchProperty(property, compareValue, false, !caseSensitive) {
		fn(d)
	}
	return d
}

// OnProfile 按 profile 条件执行：当前生效的 profile 与给定值一致时立即执行 fn。
// 与 OnProperty 同为立即求值风格；profile 解析规则见 Profile()。
func (d *dioContainer) OnProfile(profile string, fn func(core.Dio)) core.Dio {
	if d.Profile() == profile {
		fn(d)
	}
	return d
}

// OnBeanType 按 bean 类型条件执行：容器中已注册（实例/原型/工厂均可）指定类型的 bean 时立即执行 fn。
// 与 OnProperty 同为立即求值风格，可用于按已装配的 bean 类型附加注册其他 bean。
func (d *dioContainer) OnBeanType(beanType any, fn func(core.Dio)) core.Dio {
	if d.hasBeanType(beanType) {
		fn(d)
	}
	return d
}

// hasBeanType 判断容器中是否已注册指定类型的 bean。
// 直接注册进 di 的（ProvideFunc/SetLogger 等）与 dio 侧挂起的注册（Provide/RegisterBean，Run 前才同步进 di）都算。
func (d *dioContainer) hasBeanType(beanType any) bool {
	if d.di.HasBeanType(beanType) {
		return true
	}
	// 类型归一化规则与 di 一致：值类型取 *T，指针类型取 T 或 *T，接口类型保持原样
	t := reflect.TypeOf(beanType)
	if t == nil {
		return false
	}
	var typeValue reflect.Type
	if t.Kind() == reflect.Ptr {
		typeValue = t.Elem()
		if typeValue.Kind() == reflect.Struct {
			typeValue = reflect.PtrTo(typeValue)
		}
	} else {
		typeValue = reflect.PtrTo(t)
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, b := range d.providedBeans {
		// 原型/实例在容器中的实例形态都是指针（*T）
		if reflect.PtrTo(reflect.Indirect(reflect.ValueOf(b.instance)).Type()).AssignableTo(typeValue) {
			return true
		}
	}
	return false
}

func (d *dioContainer) ProvideOnProperty(prototype any, property string, compareValue string, caseSensitive ...bool) core.Dio {
	return d.ProvideNamedBeanOnProperty("", prototype, property, compareValue, caseSensitive...)
}

func (d *dioContainer) ProvideMultiBeanOnProperty(beans []any, property string, compareValue string, caseSensitive ...bool) core.Dio {
	for _, bean := range beans {
		d.ProvideOnProperty(bean, property, compareValue, caseSensitive...)
	}
	return d
}

func (d *dioContainer) ProvideNamedBeanOnProperty(beanName string, prototype any, property string, compareValue string, caseSensitive ...bool) core.Dio {
	return d.provideBeanCondition(beanName, prototype, property, compareValue, true, len(caseSensitive) > 0 && caseSensitive[0])
}

func (d *dioContainer) ProvideMultiNamedBeanOnProperty(namedBeanMap map[string]any, property string, compareValue string, caseSensitive ...bool) core.Dio {
	for name, bean := range namedBeanMap {
		d.ProvideNamedBeanOnProperty(name, bean, property, compareValue, caseSensitive...)
	}
	return d
}

func (d *dioContainer) ProvideNotOnProperty(prototype any, property string, compareValue string, caseSensitive ...bool) core.Dio {
	return d.ProvideNamedBeanNotOnProperty("", prototype, property, compareValue, caseSensitive...)
}

func (d *dioContainer) ProvideMultiBeanNotOnProperty(beans []any, property string, compareValue string, caseSensitive ...bool) core.Dio {
	for _, bean := range beans {
		d.ProvideNotOnProperty(bean, property, compareValue, caseSensitive...)
	}
	return d
}

func (d *dioContainer) ProvideNamedBeanNotOnProperty(beanName string, prototype any, property string, compareValue string, caseSensitive ...bool) core.Dio {
	return d.provideBeanCondition(beanName, prototype, property, compareValue, false, len(caseSensitive) > 0 && caseSensitive[0])
}

func (d *dioContainer) ProvideMultiNamedBeanNotOnProperty(namedBeanMap map[string]any, property string, compareValue string, caseSensitive ...bool) core.Dio {
	for name, bean := range namedBeanMap {
		d.ProvideNamedBeanNotOnProperty(name, bean, property, compareValue, caseSensitive...)
	}
	return d
}

func (d *dioContainer) GetBean(beanName string) (bean any, ok bool) {
	return d.di.GetBean(beanName)
}

func (d *dioContainer) GetByType(beanType any) (bean any, ok bool) {
	return d.di.GetByType(beanType)
}

// GetByTypeAll 按类型获取所有匹配的 bean（含名称），按注册顺序返回。
// 注意：该方法不在 core.Dio 接口中（避免接口变更），可通过 dio.GetByTypeAll 全局函数
// 或 *dioContainer 类型断言调用。Serve 退出销毁 bean 后实例已移除，返回空。
func (d *dioContainer) GetByTypeAll(beanType any) (beans []di.BeanWithName) {
	return d.di.GetByTypeAll(beanType)
}

// GetBeanNames 返回所有已注册 bean 的名称（按注册顺序，含工厂 bean）。
// 注意：不在 core.Dio 接口中（避免接口变更），可通过 dio.GetBeanNames 全局函数调用。
func (d *dioContainer) GetBeanNames() []string {
	return d.di.GetBeanNames()
}

// DescribeBean 返回 bean 定义的只读描述（供诊断/管理）。
// 仅原型（Provide）与工厂（ProvideFunc）bean 有定义；直接注册的实例（RegisterBean）返回 ok=false。
func (d *dioContainer) DescribeBean(beanName string) (desc di.BeanDescription, ok bool) {
	return d.di.DescribeBean(beanName)
}

// GetBeanDependencies 返回 bean 依赖的其他 bean 名称列表（命名 aware 注入，按名称排序）。
func (d *dioContainer) GetBeanDependencies(beanName string) (deps []string, ok bool) {
	return d.di.GetBeanDependencies(beanName)
}

func (d *dioContainer) NewBean(beanType any) (bean any) {
	return d.di.NewBean(beanType)
}

func (d *dioContainer) NewBeanByName(beanName string) (bean any) {
	return d.di.NewBeanByName(beanName)
}

func (d *dioContainer) BeanName() string {
	return "dioContainer"
}

// OnShutdown 注册优雅停机回调，容器停机阶段（Stopping）执行。
// 注意：di.Serve 退出时内部已先倒序销毁 bean（触发 Destroy 回调），OnShutdown 在 bean 销毁后执行。
// 默认按注册倒序顺序执行；可通过 SetShutdownParallel 改为并行、SetShutdownTimeout 限制总耗时。
// 可用于关闭连接池、刷新缓冲等。必须在 Run 前调用。
func (d *dioContainer) OnShutdown(fn func()) core.Dio {
	d.mu.Lock()
	d.shutdownFns = append(d.shutdownFns, fn)
	d.mu.Unlock()
	return d
}

// SetShutdownTimeout 设置停机回调（OnShutdown）的总超时时间，默认 0 表示不限时。
// 顺序模式下每个回调执行前检查时限，超时则跳过剩余回调；
// 并行模式下等待全部完成或超时（超时后不再等待，未完成的回调在后台继续执行）。
// 注意：单个回调自身阻塞无法被中断（Go 无法强制终止 goroutine）；
// 超时仅约束 OnShutdown 回调阶段，bean 销毁（di.Serve 内部）不受此限制。
func (d *dioContainer) SetShutdownTimeout(timeout time.Duration) core.Dio {
	d.mu.Lock()
	d.shutdownTimeout = timeout
	d.mu.Unlock()
	return d
}

// SetShutdownParallel 设置停机回调（OnShutdown）是否并行执行。
// 默认 false：按注册倒序依次执行；true：全部并发执行并等待完成（受 SetShutdownTimeout 约束）。
// 注意：并行 + 超时后未完成的回调继续在后台执行，但此时容器 bean 已销毁，
// 回调内不应再访问容器读 API（GetBean/GetByType 等会返回空）。
func (d *dioContainer) SetShutdownParallel(parallel bool) core.Dio {
	d.mu.Lock()
	d.shutdownParallel = parallel
	d.mu.Unlock()
	return d
}

// SetBanner 设置启动 banner，Run 开始时打印到控制台。
// 传空字符串可关闭 banner。非并发安全，必须在 Run 前调用。
func (d *dioContainer) SetBanner(banner string) core.Dio {
	d.banner = banner
	return d
}

// SetProfile 设置应用运行环境（profile），影响配置加载（LoadConfig 自动加载 config-{profile}.yaml 覆盖）
// 与条件装配（OnProfile/ProvideOnProfile）。
// 也可通过环境变量 APP_PROFILE 指定；显式 SetProfile 优先于环境变量。
// 非并发安全，必须在 Run 前调用。
func (d *dioContainer) SetProfile(profile string) core.Dio {
	d.profile = profile
	return d
}

// Profile 返回当前生效的 profile：显式 SetProfile 优先，其次环境变量 APP_PROFILE，否则为空串。
func (d *dioContainer) Profile() string {
	if d.profile != "" {
		return d.profile
	}
	return os.Getenv("APP_PROFILE")
}

// RequireProperties 声明必填配置项，Run 启动时校验：任一缺失则 panic（ErrMissingProperty，可用 errors.Is 判断）。
// 必须在 Run 前调用；配置链（LoadConfig/profile 覆盖/环境变量/显式 Set）完成后生效。
func (d *dioContainer) RequireProperties(keys ...string) core.Dio {
	d.mu.Lock()
	d.requiredProps = append(d.requiredProps, keys...)
	d.mu.Unlock()
	return d
}

// checkMissingProperties 返回缺失的必填配置项（无则返回空）。
func (d *dioContainer) checkMissingProperties() (missing []string) {
	d.mu.Lock()
	keys := append([]string{}, d.requiredProps...)
	d.mu.Unlock()
	for _, key := range keys {
		if !d.HasProperty(key) {
			missing = append(missing, key)
		}
	}
	return
}

// StartupDuration 返回启动耗时：Run 开始至今的时间。
// 未 Run 时返回 0；Running 状态时即为本次启动所用时间。
func (d *dioContainer) StartupDuration() time.Duration {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.startTime.IsZero() {
		return 0
	}
	return time.Since(d.startTime)
}

// State 返回当前应用状态。
// 未 Run 时为 Pending；Run 后按 Starting→Running→Stopping→Stopped 单向推进，失败为 Failed。
func (d *dioContainer) State() AppState {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.state
}

// OnStateChange 注册状态变更回调，容器状态推进时倒序执行（回调在锁外调用）。
// 可用于就绪探针、指标上报等。必须在 Run 前注册。
// 注意：Starting 在 bean 注册之前触发，此时容器内容不完整（仅记录启动开始事件）；
// 需要就绪感知的请用 Running 状态或 Ready()。
func (d *dioContainer) OnStateChange(fn func(AppState)) core.Dio {
	d.mu.Lock()
	d.stateChangeFns = append(d.stateChangeFns, fn)
	d.mu.Unlock()
	return d
}

// setState 推进应用状态并触发状态变更回调（快照后锁外倒序执行）。
// 状态只允许单向推进，回退或重复设置会 panic；
// 唯一例外是 Failed 后可重新 Starting——配合 Run 的 defer 还原 loaded=false 实现启动失败重试。
func (d *dioContainer) setState(s AppState) {
	d.mu.Lock()
	if s <= d.state && !(d.state == Failed && s == Starting) {
		old := d.state
		d.mu.Unlock()
		panic(fmt.Errorf("dio state cannot go backwards: %s -> %s", old, s))
	}
	d.state = s
	fns := append([]func(AppState){}, d.stateChangeFns...)
	d.mu.Unlock()
	// 回调在锁外倒序执行（回调可能反向调用容器方法，持锁会死锁）
	for i := len(fns) - 1; i >= 0; i-- {
		fns[i](s)
	}
}

// Ready 返回容器是否就绪：应用状态为 Running（已 Run 且 afterRunFns 执行完毕）。
// 可用于就绪探针（如 /health 检查），避免在服务未完全启动时接收流量。
func (d *dioContainer) Ready() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.state == Running
}

// Health 健康检查：聚合容器内所有实现 core.HealthChecker 的 bean 的检查结果。
// 每个检查独立超时（healthCheckTimeout），单个失败即返回聚合错误（errors.Join，可用 errors.Is 判断）。
// 容器未就绪（Ready=false）时直接返回错误，避免在启动未完成时误报健康。
func (d *dioContainer) Health(ctx context.Context) error {
	if !d.Ready() {
		return errors.New("dio not ready")
	}
	checkers := d.di.GetByTypeAll((*core.HealthChecker)(nil))
	var errs []error
	for _, checkerBean := range checkers {
		// 容器自身注册为 bean 且实现 HealthChecker（Health 即聚合入口），跳过以避免无限递归
		if checkerBean.Bean == d {
			continue
		}
		checker := checkerBean.Bean.(core.HealthChecker)
		// 每个检查独立超时；ctx 若带 deadline 会进一步收紧
		checkCtx, cancel := context.WithTimeout(ctx, healthCheckTimeout)
		err := checker.Health(checkCtx)
		cancel()
		if err != nil {
			errs = append(errs, fmt.Errorf("health check failed for %s: %w", checkerBean.Name, err))
		}
	}
	return errors.Join(errs...)
}

func (d *dioContainer) Run(ctx context.Context, afterRunFns ...func(core.Dio)) {
	if d.loaded {
		panic(fmt.Errorf("%w: dioContainer is already run", ErrAlreadyRun))
	}
	d.loaded = true
	d.mu.Lock()
	d.startTime = time.Now()
	d.mu.Unlock()

	// panic 时还原 loaded 并关闭已创建的日志组件，避免状态残留与资源泄漏。
	// 失败时状态直接置 Failed（不经 setState：回调在 panic 展开期执行可能再次 panic 掩盖原始错误），
	// 调用方可通过 State() 感知失败。
	// 注意：仅日志创建阶段的失败可修正后重试 Run；
	// bean 注册/di.Load 之后的失败，di 容器已残留 bean 与 loaded 状态，重试会 panic。
	defer func() {
		if r := recover(); r != nil {
			d.loaded = false
			d.mu.Lock()
			d.state = Failed
			d.mu.Unlock()
			if d.log != nil {
				if disposable, ok := d.log.(Disposable); ok {
					_ = disposable.Close()
				}
			}
			panic(r)
		}
	}()

	d.setState(Starting)

	// 打印启动 banner（与日志组件无关，直接输出到控制台）
	if d.banner != "" {
		fmt.Println(d.banner)
	}

	// 必填配置项校验：缺失则启动失败。放在日志创建之前，失败不污染 di 容器，修正后可重试
	if missing := d.checkMissingProperties(); len(missing) > 0 {
		panic(fmt.Errorf("%w: %s", ErrMissingProperty, strings.Join(missing, ", ")))
	}

	// 先创建日志组件再注册容器 bean：日志创建阶段的失败不污染 di 容器（未注册任何 bean），
	// 修正后可重试 Run；bean 注册/di.Load 之后的失败，di 容器已残留 bean，重试会 panic。
	if d.log == nil {
		property := d.GetProperties("log.", core.Property{}).(core.Property)
		if log, err := NewZapLogger(property); err != nil {
			panic(err)
		} else {
			d.log = log
		}
		d.di.RegisterBean(d.log)
	}
	d.di.RegisterBean(d)
	d.di.Log(newDiLogger(d.log))

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 配置bean（锁内取快照，避免与并发注册 race）
	d.mu.Lock()
	providedBeans := append([]bean(nil), d.providedBeans...)
	d.mu.Unlock()
	for _, beanDefinition := range providedBeans {
		if beanDefinition.matchProperty(d) {
			if beanDefinition.registered {
				d.di.RegisterNamedBean(beanDefinition.name, beanDefinition.instance)
			} else {
				d.di.ProvideNamedBean(beanDefinition.name, beanDefinition.instance)
			}
		}
	}

	// 启动容器
	d.di.Load()

	// 容器加载完成后执行的方法
	for _, fn := range afterRunFns {
		fn(d)
	}
	// 进入 Running 状态（触发 OnStateChange 回调），并输出启动摘要（bean 数/耗时/profile）
	d.setState(Running)
	summary := fmt.Sprintf("started in %s, %d beans", d.StartupDuration(), len(d.di.GetBeanNames()))
	if profile := d.Profile(); profile != "" {
		summary += fmt.Sprintf(", profile: %s", profile)
	}
	d.log.Info(context.Background(), summary)

	// 阻塞等待 ctx 结束；di.Serve 退出时内部已倒序销毁 bean（触发 Destroy 回调）
	d.di.Serve(ctx)

	// Serve 退出：进入停机阶段，执行停机回调（bean 已在 di.Serve 内部销毁）
	d.setState(Stopping)
	d.mu.Lock()
	shutdownFns := append([]func(){}, d.shutdownFns...)
	shutdownTimeout := d.shutdownTimeout
	shutdownParallel := d.shutdownParallel
	d.mu.Unlock()
	d.runShutdownFns(shutdownFns, shutdownTimeout, shutdownParallel)
	d.setState(Stopped)
}

// runShutdownFns 执行停机回调（倒序）。
// parallel 为 true 时并发执行；timeout > 0 时受整体时限约束（语义见 SetShutdownTimeout）。
func (d *dioContainer) runShutdownFns(fns []func(), timeout time.Duration, parallel bool) {
	if len(fns) == 0 {
		return
	}
	var deadline time.Time
	if timeout > 0 {
		deadline = time.Now().Add(timeout)
	}
	if !parallel {
		// 顺序执行：每个回调前检查时限，超时则跳过剩余回调
		for i := len(fns) - 1; i >= 0; i-- {
			if !deadline.IsZero() && time.Now().After(deadline) {
				return
			}
			fns[i]()
		}
		return
	}
	// 并行执行：等待全部完成或整体时限到（超时后不再等待，未完成回调在后台继续执行）
	var wg sync.WaitGroup
	for i := len(fns) - 1; i >= 0; i-- {
		wg.Add(1)
		go func(fn func()) {
			defer wg.Done()
			fn()
		}(fns[i])
	}
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	if deadline.IsZero() {
		<-done
		return
	}
	select {
	case <-done:
	case <-time.After(time.Until(deadline)):
	}
}

func (d *dioContainer) Use(plugins ...core.PluginConfig) core.Dio {
	for i := range plugins {
		plugins[i](d)
	}
	return d
}

// loadConfigMap 读取并解析 yaml 配置文件为 map。文件不存在或读取失败返回错误（由调用方决定是否 panic）。
func loadConfigMap(configs fs.FS, filename string) (configMap map[string]any, err error) {
	f, err := configs.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	configMap = map[string]any{}
	if err := yaml.Unmarshal(data, &configMap); err != nil {
		return nil, err
	}
	return configMap, nil
}

// profileConfigFilename 生成 profile 覆盖配置文件名：config.yaml + profile=dev → config-dev.yaml
func profileConfigFilename(filename string, profile string) string {
	ext := path.Ext(filename)
	return strings.TrimSuffix(filename, ext) + "-" + profile + ext
}

func (d *dioContainer) LoadDefaultConfig(configs fs.FS, filename string) core.Dio {
	configMap, err := loadConfigMap(configs, filename)
	if err != nil {
		panic(err)
	}
	d.SetDefaultPropertyMap(configMap)
	return d
}

// 配置优先级链（低 → 高；同级内后写覆盖先写）：
//   - SetDefaultProperty/SetDefaultPropertyMap/LoadDefaultConfig：默认配置
//   - LoadConfig/LoadConfigDir 的公共配置文件：与默认配置同级，加载在后覆盖同名项
//   - LoadConfig 的 config-{profile}.yaml 覆盖配置：高于公共配置文件
//   - AutoMigrateEnv 环境变量、SetProperty/SetPropertyMap 显式配置：最高优先级（同级，后写覆盖先写）

func (d *dioContainer) LoadConfig(configs fs.FS, filename string) core.Dio {
	// 公共配置为低优先级（SetDefault），profile 覆盖配置为高优先级（Set），优先级链见上方注释
	configMap, err := loadConfigMap(configs, filename)
	if err != nil {
		panic(err)
	}
	d.SetDefaultPropertyMap(configMap)
	// 自动尝试加载 profile 覆盖配置（如 config-dev.yaml），文件不存在时忽略
	if profile := d.Profile(); profile != "" {
		profileConfigMap, err := loadConfigMap(configs, profileConfigFilename(filename, profile))
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				panic(err)
			}
		} else {
			d.SetPropertyMap(profileConfigMap)
		}
	}
	return d
}

// LoadConfigDir 加载目录下所有 *.yaml 配置文件（按文件名排序，后加载的覆盖同名配置项），
// 全部作为公共配置（SetDefault 优先级）合并。目录不存在或文件解析失败会 panic。
// 注意：目录加载不区分 profile，按 profile 覆盖请使用 LoadConfig 的 config-{profile}.yaml 约定。
func (d *dioContainer) LoadConfigDir(configs fs.FS, dir string) core.Dio {
	entries, err := fs.ReadDir(configs, dir)
	if err != nil {
		panic(err)
	}
	filenames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		filenames = append(filenames, entry.Name())
	}
	sort.Strings(filenames)
	for _, filename := range filenames {
		configMap, err := loadConfigMap(configs, path.Join(dir, filename))
		if err != nil {
			panic(err)
		}
		d.SetDefaultPropertyMap(configMap)
	}
	return d
}
