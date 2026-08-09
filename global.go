package dio

import (
	"context"
	"io/fs"
	"sync"

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

func Run(ctx context.Context) {
	container().Run(ctx)
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
