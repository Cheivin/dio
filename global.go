package dio

import (
	"context"
	"io/fs"
	"sync"
	"time"

	"github.com/cheivin/di"
	"github.com/cheivin/dio-core"
)

var (
	gMu sync.Mutex
	g   core.Dio
)

// container 懒初始化并返回全局容器实例。
// 未调用任何全局函数前不创建容器，首次访问时创建。
func container() core.Dio {
	gMu.Lock()
	defer gMu.Unlock()
	if g == nil {
		g = New()
	}
	return g
}

// Reset 将全局容器重置为未初始化状态（清空所有 bean 与配置，下次调用时懒创建）。
// 仅用于测试隔离（全局容器有状态残留），生产代码不应调用。
func Reset() {
	gMu.Lock()
	defer gMu.Unlock()
	g = nil
}

func SetDefaultProperty(key string, value any) core.Dio {
	return container().SetDefaultProperty(key, value)
}

func SetDefaultPropertyMap(properties map[string]any) core.Dio {
	return container().SetDefaultPropertyMap(properties)
}

func SetProperty(key string, value any) core.Dio {
	return container().SetProperty(key, value)
}

func GetPropertyString(key string) string {
	return container().GetPropertyString(key)
}

func GetProperties(prefix string, destType any) any {
	return container().GetProperties(prefix, destType)
}

func SetPropertyMap(properties map[string]any) core.Dio {
	return container().SetPropertyMap(properties)
}

func AutoMigrateEnv() core.Dio {
	return container().AutoMigrateEnv()
}

func SetLogger(log core.Log) core.Dio {
	return container().SetLogger(log)
}

func Logger() core.Log {
	return container().Logger()
}

func RegisterBean(bean any) core.Dio {
	return container().RegisterBean(bean)
}

func RegisterNamedBean(name string, bean any) core.Dio {
	return container().RegisterNamedBean(name, bean)
}

func Provide(prototype ...any) core.Dio {
	return container().Provide(prototype...)
}

func ProvideNamedBean(beanName string, prototype any) core.Dio {
	return container().ProvideNamedBean(beanName, prototype)
}

func ProvideMultiNamedBean(namedBeanMap map[string]any) core.Dio {
	return container().ProvideMultiNamedBean(namedBeanMap)
}

func ProvideFunc(fn any) core.Dio {
	return container().ProvideFunc(fn)
}

// WithCircularCheck 开启/关闭循环依赖检测（默认关闭）
func WithCircularCheck(enable bool) core.Dio {
	return container().WithCircularCheck(enable)
}

// WithBeanSelector 设置接口多实现时的选择策略。
// 需要引入 github.com/cheivin/di 包的 BeanSelector 类型，故以独立函数提供。
func WithBeanSelector(s di.BeanSelector) core.Dio {
	c := container()
	if dc, ok := c.(*dioContainer); ok {
		dc.WithBeanSelector(s)
	}
	return c
}

func ProvideOnProperty(prototype any, property string, compareValue string, caseSensitive ...bool) core.Dio {
	return container().ProvideOnProperty(prototype, property, compareValue, caseSensitive...)
}

func ProvideNamedBeanOnProperty(beanName string, prototype any, property string, compareValue string, caseSensitive ...bool) core.Dio {
	return container().ProvideNamedBeanOnProperty(beanName, prototype, property, compareValue, caseSensitive...)
}

func ProvideMultiBeanOnProperty(beans []any, property string, compareValue string, caseSensitive ...bool) core.Dio {
	return container().ProvideMultiBeanOnProperty(beans, property, compareValue, caseSensitive...)
}
func ProvideMultiNamedBeanOnProperty(namedBeanMap map[string]any, property string, compareValue string, caseSensitive ...bool) core.Dio {
	return container().ProvideMultiNamedBeanOnProperty(namedBeanMap, property, compareValue, caseSensitive...)
}

func ProvideNotOnProperty(prototype any, property string, compareValue string, caseSensitive ...bool) core.Dio {
	return container().ProvideNotOnProperty(prototype, property, compareValue, caseSensitive...)
}

func ProvideNamedBeanNotOnProperty(beanName string, prototype any, property string, compareValue string, caseSensitive ...bool) core.Dio {
	return container().ProvideNamedBeanNotOnProperty(beanName, prototype, property, compareValue, caseSensitive...)
}

func ProvideMultiBeanNotOnProperty(beans []any, property string, compareValue string, caseSensitive ...bool) core.Dio {
	return container().ProvideMultiBeanNotOnProperty(beans, property, compareValue, caseSensitive...)
}
func ProvideMultiNamedBeanNotOnProperty(namedBeanMap map[string]any, property string, compareValue string, caseSensitive ...bool) core.Dio {
	return container().ProvideMultiNamedBeanNotOnProperty(namedBeanMap, property, compareValue, caseSensitive...)
}

func GetBean(beanName string) (bean any, ok bool) {
	return container().GetBean(beanName)
}

func GetByType(beanType any) (bean any, ok bool) {
	return container().GetByType(beanType)
}

// GetByTypeAll 按类型获取所有匹配的 bean。
func GetByTypeAll(beanType any) (beans []di.BeanWithName) {
	return container().(*dioContainer).GetByTypeAll(beanType)
}

// GetBeanNames 返回所有已注册 bean 的名称（按注册顺序）。
func GetBeanNames() []string {
	return container().(*dioContainer).GetBeanNames()
}

// DescribeBean 返回 bean 定义的只读描述（仅原型/工厂 bean 有定义）。
func DescribeBean(beanName string) (desc di.BeanDescription, ok bool) {
	return container().(*dioContainer).DescribeBean(beanName)
}

// GetBeanDependencies 返回 bean 依赖的其他 bean 名称列表。
func GetBeanDependencies(beanName string) (deps []string, ok bool) {
	return container().(*dioContainer).GetBeanDependencies(beanName)
}

func HasProperty(property string) bool {
	return container().HasProperty(property)
}

func NewBean(beanType any) (bean any) {
	return container().NewBean(beanType)
}

func NewBeanByName(beanName string) (bean any) {
	return container().NewBeanByName(beanName)
}

func OnProperty(property string, compareValue string, caseSensitive bool, fn func(core.Dio)) core.Dio {
	return container().OnProperty(property, compareValue, caseSensitive, fn)
}

func NotOnProperty(property string, compareValue string, caseSensitive bool, fn func(core.Dio)) core.Dio {
	return container().NotOnProperty(property, compareValue, caseSensitive, fn)
}

// OnProfile 按 profile 条件执行（立即求值）。
func OnProfile(profile string, fn func(core.Dio)) core.Dio {
	return container().(*dioContainer).OnProfile(profile, fn)
}

// OnBeanType 按 bean 类型条件执行（立即求值）。
func OnBeanType(beanType any, fn func(core.Dio)) core.Dio {
	return container().(*dioContainer).OnBeanType(beanType, fn)
}

func Run(ctx context.Context) {
	container().Run(ctx)
}

// SetBanner 设置启动 banner（传空字符串关闭）。
func SetBanner(banner string) core.Dio {
	return container().(*dioContainer).SetBanner(banner)
}

// StartupDuration 返回全局容器启动耗时（未 Run 时返回 0）。
func StartupDuration() time.Duration {
	return container().(*dioContainer).StartupDuration()
}

// SetProfile 设置应用运行环境（profile），影响配置加载与条件装配。
func SetProfile(profile string) core.Dio {
	return container().(*dioContainer).SetProfile(profile)
}

// Profile 返回当前生效的 profile（显式 SetProfile 优先，其次环境变量 APP_PROFILE）。
func Profile() string {
	return container().(*dioContainer).Profile()
}

// RequireProperties 声明必填配置项，Run 启动时校验缺失则 panic（ErrMissingProperty）。
func RequireProperties(keys ...string) core.Dio {
	return container().(*dioContainer).RequireProperties(keys...)
}

// OnShutdown 注册优雅停机回调，Run 停机阶段执行（bean 销毁后，默认倒序顺序执行）。
func OnShutdown(fn func()) core.Dio {
	return container().(*dioContainer).OnShutdown(fn)
}

// SetShutdownTimeout 设置停机回调总超时（0 表示不限时）。
func SetShutdownTimeout(timeout time.Duration) core.Dio {
	return container().(*dioContainer).SetShutdownTimeout(timeout)
}

// SetShutdownParallel 设置停机回调是否并行执行（默认顺序倒序）。
func SetShutdownParallel(parallel bool) core.Dio {
	return container().(*dioContainer).SetShutdownParallel(parallel)
}

// State 返回全局容器当前应用状态。
func State() AppState {
	return container().(*dioContainer).State()
}

// OnStateChange 注册全局容器状态变更回调，状态推进时倒序执行。
func OnStateChange(fn func(AppState)) core.Dio {
	return container().(*dioContainer).OnStateChange(fn)
}

// Ready 返回全局容器是否就绪。
func Ready() bool {
	return container().(*dioContainer).Ready()
}

// Health 健康检查：聚合容器内所有 core.HealthChecker bean 的检查结果。
func Health(ctx context.Context) error {
	return container().(*dioContainer).Health(ctx)
}

func Use(plugins ...core.PluginConfig) core.Dio {
	return container().Use(plugins...)
}

func LoadDefaultConfig(configs fs.FS, filename string) core.Dio {
	return container().LoadDefaultConfig(configs, filename)
}

func LoadConfig(configs fs.FS, filename string) core.Dio {
	return container().LoadConfig(configs, filename)
}

// LoadConfigDir 加载目录下所有 *.yaml 配置文件（按文件名排序合并，后加载覆盖同名项）。
func LoadConfigDir(configs fs.FS, dir string) core.Dio {
	return container().(*dioContainer).LoadConfigDir(configs, dir)
}
