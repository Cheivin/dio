package dio

import (
	"context"
	"gorm.io/gorm"
)

func SetDefaultProperty(key string, value interface{}) *dio {
	return g.SetDefaultProperty(key, value)
}

func SetDefaultPropertyMap(properties map[string]interface{}) *dio {
	return g.SetDefaultPropertyMap(properties)
}

func SetProperty(key string, value interface{}) *dio {
	return g.SetProperty(key, value)
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

func Provide(prototype interface{}) *dio {
	return g.Provide(prototype)
}

func ProvideWithBeanName(beanName string, prototype interface{}) *dio {
	return g.ProvideWithBeanName(beanName, prototype)
}

func GetBean(beanName string) (bean interface{}, ok bool) {
	return g.GetBean(beanName)
}

func Run(ctx context.Context) {
	g.Run(ctx)
}

func Web(useLogger, useCors bool) *dio {
	return g.Web(useLogger, useCors)
}

func MySQL(options ...gorm.Option) *dio {
	return g.MySQL(options...)
}
