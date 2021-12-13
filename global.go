package dio

import (
	"context"
	"embed"
	"github.com/cheivin/dio-core"
)

var g dio.Dio

func init() {
	g = New()
}

func SetDefaultProperty(key string, value interface{}) dio.Dio {
	return g.SetDefaultProperty(key, value)
}

func SetDefaultPropertyMap(properties map[string]interface{}) dio.Dio {
	return g.SetDefaultPropertyMap(properties)
}

func SetProperty(key string, value interface{}) dio.Dio {
	return g.SetProperty(key, value)
}

func GetPropertyString(key string) string {
	return g.GetPropertyString(key)
}

func SetPropertyMap(properties map[string]interface{}) dio.Dio {
	return g.SetPropertyMap(properties)
}

func AutoMigrateEnv() dio.Dio {
	return g.AutoMigrateEnv()
}

func RegisterBean(bean interface{}) dio.Dio {
	return g.RegisterBean(bean)
}

func RegisterNamedBean(name string, bean interface{}) dio.Dio {
	return g.RegisterNamedBean(name, bean)
}

func Provide(prototype ...interface{}) dio.Dio {
	return g.Provide(prototype...)
}

func ProvideNamedBean(beanName string, prototype interface{}) dio.Dio {
	return g.ProvideNamedBean(beanName, prototype)
}

func ProvideMultiNamedBean(namedBeanMap map[string]interface{}) dio.Dio {
	return g.ProvideMultiNamedBean(namedBeanMap)
}

func ProvideOnProperty(prototype interface{}, property string, compareValue string, caseSensitive ...bool) dio.Dio {
	return g.ProvideOnProperty(prototype, property, compareValue, caseSensitive...)
}

func ProvideNamedBeanOnProperty(beanName string, prototype interface{}, property string, compareValue string, caseSensitive ...bool) dio.Dio {
	return g.ProvideNamedBeanOnProperty(beanName, prototype, property, compareValue, caseSensitive...)
}

func ProvideMultiBeanOnProperty(beans []interface{}, property string, compareValue string, caseSensitive ...bool) dio.Dio {
	return g.ProvideMultiBeanOnProperty(beans, property, compareValue, caseSensitive...)
}
func ProvideMultiNamedBeanOnProperty(namedBeanMap map[string]interface{}, property string, compareValue string, caseSensitive ...bool) dio.Dio {
	return g.ProvideMultiNamedBeanOnProperty(namedBeanMap, property, compareValue, caseSensitive...)
}

func ProvideNotOnProperty(prototype interface{}, property string, compareValue string, caseSensitive ...bool) dio.Dio {
	return g.ProvideNotOnProperty(prototype, property, compareValue, caseSensitive...)
}

func ProvideNamedBeanNotOnProperty(beanName string, prototype interface{}, property string, compareValue string, caseSensitive ...bool) dio.Dio {
	return g.ProvideNamedBeanNotOnProperty(beanName, prototype, property, compareValue, caseSensitive...)
}

func ProvideMultiBeanNotOnProperty(beans []interface{}, property string, compareValue string, caseSensitive ...bool) dio.Dio {
	return g.ProvideMultiBeanNotOnProperty(beans, property, compareValue, caseSensitive...)
}
func ProvideMultiNamedBeanNotOnProperty(namedBeanMap map[string]interface{}, property string, compareValue string, caseSensitive ...bool) dio.Dio {
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

func Use(plugins ...dio.PluginConfig) dio.Dio {
	return g.Use(plugins...)
}

func LoadDefaultConfig(configs embed.FS, filename string) dio.Dio {
	return g.LoadDefaultConfig(configs, filename)
}

func LoadConfig(configs embed.FS, filename string) dio.Dio {
	return g.LoadConfig(configs, filename)
}
