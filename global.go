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
		"trace-name": DefaultTraceName,
	})
}

func SetDefaultProperty(key string, value interface{}) Dio {
	return g.SetDefaultProperty(key, value)
}

func SetDefaultPropertyMap(properties map[string]interface{}) Dio {
	return g.SetDefaultPropertyMap(properties)
}

func SetProperty(key string, value interface{}) Dio {
	return g.SetProperty(key, value)
}

func GetPropertyString(key string) string {
	return g.GetPropertyString(key)
}

func SetPropertyMap(properties map[string]interface{}) Dio {
	return g.SetPropertyMap(properties)
}

func AutoMigrateEnv() Dio {
	return g.AutoMigrateEnv()
}

func RegisterBean(bean interface{}) Dio {
	return g.RegisterBean(bean)
}

func RegisterNamedBean(name string, bean interface{}) Dio {
	return g.RegisterNamedBean(name, bean)
}

func Provide(prototype ...interface{}) Dio {
	return g.Provide(prototype...)
}

func ProvideNamedBean(beanName string, prototype interface{}) Dio {
	return g.ProvideNamedBean(beanName, prototype)
}

func ProvideMultiNamedBean(namedBeanMap map[string]interface{}) Dio {
	return g.ProvideMultiNamedBean(namedBeanMap)
}

func ProvideOnProperty(prototype interface{}, property string, compareValue string, caseSensitive ...bool) Dio {
	return g.ProvideOnProperty(prototype, property, compareValue, caseSensitive...)
}

func ProvideNamedBeanOnProperty(beanName string, prototype interface{}, property string, compareValue string, caseSensitive ...bool) Dio {
	return g.ProvideNamedBeanOnProperty(beanName, prototype, property, compareValue, caseSensitive...)
}

func ProvideMultiBeanOnProperty(beans []interface{}, property string, compareValue string, caseSensitive ...bool) Dio {
	return g.ProvideMultiBeanOnProperty(beans, property, compareValue, caseSensitive...)
}
func ProvideMultiNamedBeanOnProperty(namedBeanMap map[string]interface{}, property string, compareValue string, caseSensitive ...bool) Dio {
	return g.ProvideMultiNamedBeanOnProperty(namedBeanMap, property, compareValue, caseSensitive...)
}

func ProvideNotOnProperty(prototype interface{}, property string, compareValue string, caseSensitive ...bool) Dio {
	return g.ProvideNotOnProperty(prototype, property, compareValue, caseSensitive...)
}

func ProvideNamedBeanNotOnProperty(beanName string, prototype interface{}, property string, compareValue string, caseSensitive ...bool) Dio {
	return g.ProvideNamedBeanNotOnProperty(beanName, prototype, property, compareValue, caseSensitive...)
}

func ProvideMultiBeanNotOnProperty(beans []interface{}, property string, compareValue string, caseSensitive ...bool) Dio {
	return g.ProvideMultiBeanNotOnProperty(beans, property, compareValue, caseSensitive...)
}
func ProvideMultiNamedBeanNotOnProperty(namedBeanMap map[string]interface{}, property string, compareValue string, caseSensitive ...bool) Dio {
	return g.ProvideMultiNamedBeanNotOnProperty(namedBeanMap, property, compareValue, caseSensitive...)
}

func GetBean(beanName string) (bean interface{}, ok bool) {
	return g.GetBean(beanName)
}

func GetByType(beanType interface{}) (bean interface{}, ok bool) {
	return g.GetByType(beanType)
}

func Run(ctx context.Context) {
	g.Run(ctx)
}

func Use(plugins ...PluginConfig) Dio {
	return g.Use(plugins...)
}

func LoadDefaultConfig(configs embed.FS, filename string) Dio {
	return g.LoadDefaultConfig(configs, filename)
}

func LoadConfig(configs embed.FS, filename string) Dio {
	return g.LoadConfig(configs, filename)
}
