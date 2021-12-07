package dio

import (
	"context"
	"embed"
	"github.com/cheivin/di"
	"os"
)

var g *dio

func init() {
	g = &dio{di: di.New(), providedBeans: []bean{}, loaded: false}
	g.di.Log(emptyLogger{})
	logName := "dio_app"
	if hostname, err := os.Hostname(); err == nil && hostname != "" {
		logName += "_" + hostname
	}
	g.SetDefaultProperty("log", map[string]interface{}{
		"name":       logName,
		"dir":        "./logs",
		"max-age":    30,
		"debug":      true,
		"std":        true,
		"file":       true,
		"trace-name": defaultTraceName,
	})
}

func SetDefaultProperty(key string, value interface{}) *dio {
	return g.SetDefaultProperty(key, value)
}

func SetDefaultPropertyMap(properties map[string]interface{}) *dio {
	return g.SetDefaultPropertyMap(properties)
}

func SetProperty(key string, value interface{}) *dio {
	return g.SetProperty(key, value)
}

func GetPropertyString(key string) string {
	return g.GetPropertyString(key)
}

func SetPropertyMap(properties map[string]interface{}) *dio {
	return g.SetPropertyMap(properties)
}

func AutoMigrateEnv() *dio {
	return g.AutoMigrateEnv()
}

func RegisterBean(bean interface{}) *dio {
	return g.RegisterBean(bean)
}

func RegisterNamedBean(name string, bean interface{}) *dio {
	return g.RegisterNamedBean(name, bean)
}

func Provide(prototype ...interface{}) *dio {
	return g.Provide(prototype...)
}

func ProvideNamedBean(beanName string, prototype interface{}) *dio {
	return g.ProvideNamedBean(beanName, prototype)
}

func ProvideMultiNamedBean(namedBeanMap map[string]interface{}) *dio {
	return g.ProvideMultiNamedBean(namedBeanMap)
}

func ProvideOnProperty(prototype interface{}, property string, compareValue string, caseSensitive ...bool) *dio {
	return g.ProvideOnProperty(prototype, property, compareValue, caseSensitive...)
}

func ProvideNamedBeanOnProperty(beanName string, prototype interface{}, property string, compareValue string, caseSensitive ...bool) *dio {
	return g.ProvideNamedBeanOnProperty(beanName, prototype, property, compareValue, caseSensitive...)
}

func ProvideMultiBeanOnProperty(beans []interface{}, property string, compareValue string, caseSensitive ...bool) *dio {
	return g.ProvideMultiBeanOnProperty(beans, property, compareValue, caseSensitive...)
}
func ProvideMultiNamedBeanOnProperty(namedBeanMap map[string]interface{}, property string, compareValue string, caseSensitive ...bool) *dio {
	return g.ProvideMultiNamedBeanOnProperty(namedBeanMap, property, compareValue, caseSensitive...)
}

func ProvideNotOnProperty(prototype interface{}, property string, compareValue string, caseSensitive ...bool) *dio {
	return g.ProvideNotOnProperty(prototype, property, compareValue, caseSensitive...)
}

func ProvideNamedBeanNotOnProperty(beanName string, prototype interface{}, property string, compareValue string, caseSensitive ...bool) *dio {
	return g.ProvideNamedBeanNotOnProperty(beanName, prototype, property, compareValue, caseSensitive...)
}

func ProvideMultiBeanNotOnProperty(beans []interface{}, property string, compareValue string, caseSensitive ...bool) *dio {
	return g.ProvideMultiBeanNotOnProperty(beans, property, compareValue, caseSensitive...)
}
func ProvideMultiNamedBeanNotOnProperty(namedBeanMap map[string]interface{}, property string, compareValue string, caseSensitive ...bool) *dio {
	return g.ProvideMultiNamedBeanNotOnProperty(namedBeanMap, property, compareValue, caseSensitive...)
}

func GetBean(beanName string) (bean interface{}, ok bool) {
	return g.GetBean(beanName)
}

func Run(ctx context.Context) {
	g.Run(ctx)
}

func Use(plugins ...PluginConfig) *dio {
	return g.Use(plugins...)
}

func LoadDefaultConfig(configs embed.FS, filename string) *dio {
	return g.LoadDefaultConfig(configs, filename)
}

func LoadConfig(configs embed.FS, filename string) *dio {
	return g.LoadConfig(configs, filename)
}
