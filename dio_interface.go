package dio

import (
	"context"
	"embed"
)

type Dio interface {
	// SetDefaultProperty 设置默认配置项
	SetDefaultProperty(key string, value interface{}) Dio

	// SetDefaultPropertyMap 设置多条默认配置项
	SetDefaultPropertyMap(properties map[string]interface{}) Dio

	// SetProperty 设置配置项
	SetProperty(key string, value interface{}) Dio

	// SetPropertyMap 设置多条配置项
	SetPropertyMap(properties map[string]interface{}) Dio

	// HasProperty 判断是否存在配置项
	HasProperty(property string) bool

	// GetPropertyString 获取配置项值
	GetPropertyString(property string) string

	// LoadDefaultConfig 从文件中加载默认配置
	LoadDefaultConfig(configs embed.FS, filename string) Dio

	// LoadConfig 从文件中加载配置
	LoadConfig(configs embed.FS, filename string) Dio

	// AutoMigrateEnv 载入环境变量到配置
	AutoMigrateEnv() Dio

	// RegisterBean 注册bean实例
	RegisterBean(beanInstance ...interface{}) Dio

	// RegisterNamedBean 指定名称注册bean实例
	RegisterNamedBean(beanName string, beanInstance interface{}) Dio

	// Provide 注册bean原型
	Provide(prototype ...interface{}) Dio

	// ProvideNamedBean 指定名称注册bean原型
	ProvideNamedBean(beanName string, prototype interface{}) Dio

	// ProvideMultiNamedBean 根据map注册多个bean原型
	ProvideMultiNamedBean(namedBeanMap map[string]interface{}) Dio

	// OnProperty 按条件执行
	OnProperty(property string, compareValue string, caseSensitive bool, fn func(Dio)) Dio

	// NotOnProperty 按条件执行
	NotOnProperty(property string, compareValue string, caseSensitive bool, fn func(Dio)) Dio

	// ProvideOnProperty 按条件注册bean原型
	ProvideOnProperty(prototype interface{}, property string, compareValue string, caseSensitive ...bool) Dio

	// ProvideMultiBeanOnProperty 按条件注册多个bean原型
	ProvideMultiBeanOnProperty(beans []interface{}, property string, compareValue string, caseSensitive ...bool) Dio

	// ProvideNamedBeanOnProperty 按条件指定名称注册bean原型
	ProvideNamedBeanOnProperty(beanName string, prototype interface{}, property string, compareValue string, caseSensitive ...bool) Dio

	// ProvideMultiNamedBeanOnProperty 按条件根据map注册多个bean原型
	ProvideMultiNamedBeanOnProperty(namedBeanMap map[string]interface{}, property string, compareValue string, caseSensitive ...bool) Dio

	// ProvideNotOnProperty 按条件注册bean原型
	ProvideNotOnProperty(prototype interface{}, property string, compareValue string, caseSensitive ...bool) Dio

	// ProvideMultiBeanNotOnProperty 按条件注册多个bean原型
	ProvideMultiBeanNotOnProperty(beans []interface{}, property string, compareValue string, caseSensitive ...bool) Dio

	// ProvideNamedBeanNotOnProperty 按条件指定名称注册bean原型
	ProvideNamedBeanNotOnProperty(beanName string, prototype interface{}, property string, compareValue string, caseSensitive ...bool) Dio

	// ProvideMultiNamedBeanNotOnProperty 按条件根据map注册多个bean原型
	ProvideMultiNamedBeanNotOnProperty(namedBeanMap map[string]interface{}, property string, compareValue string, caseSensitive ...bool) Dio

	// GetBean 根据名称从容器中获取bean实例
	GetBean(beanName string) (bean interface{}, ok bool)

	// GetByType 根据类型从容器中获取bean实例
	GetByType(beanType interface{}) (bean interface{}, ok bool)

	// Use 使用插件
	Use(plugins ...PluginConfig) Dio

	// Run 运行
	Run(ctx context.Context)
}

type PluginConfig func(Dio)

const DefaultTraceName = "X-Request-Id"
