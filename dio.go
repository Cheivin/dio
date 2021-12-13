package dio

import (
	"context"
	"embed"
	"fmt"
	"github.com/cheivin/di"
	"github.com/cheivin/dio/system"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os/signal"
	"strings"
	"syscall"
)

type dio struct {
	di            di.DI
	providedBeans []bean
	loaded        bool
}

type bean struct {
	name            string      // 名称
	instance        interface{} // 实例
	needMatch       bool        // 是否条件载入
	property        string      // 条件载入配置项
	compareValue    string      // 条件载入配置比较值
	caseInsensitive bool        // 条件载入配置比较值大小写敏感
	registered      bool        // 是否为手动注册的bean
}

func (b bean) matchProperty(d *dio) (match bool) {
	return d.matchProperty(b.property, b.compareValue, b.needMatch, b.caseInsensitive)
}

func (d *dio) matchProperty(property string, compareValue string, needMatch bool, caseInsensitive bool) (match bool) {
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
	if needMatch {
		return match
	} else {
		return !match
	}
}

func (d *dio) SetDefaultProperty(key string, value interface{}) Dio {
	d.di.SetDefaultProperty(key, value)
	return d
}

func (d *dio) SetDefaultPropertyMap(properties map[string]interface{}) Dio {
	d.di.SetDefaultPropertyMap(properties)
	return d
}

func (d *dio) SetProperty(key string, value interface{}) Dio {
	d.di.SetProperty(key, value)
	return d
}

func (d *dio) SetPropertyMap(properties map[string]interface{}) Dio {
	d.di.SetPropertyMap(properties)
	return d
}

func (d *dio) HasProperty(property string) bool {
	return d.di.Property().Get(property) != nil
}

func (d *dio) GetPropertyString(property string) string {
	val := d.di.Property().Get(property)
	if val == nil {
		return ""
	} else {
		return fmt.Sprintf("%v", val)
	}
}

func (d *dio) AutoMigrateEnv() Dio {
	envMap := di.LoadEnvironment(strings.NewReplacer("_", "."), false)
	d.SetPropertyMap(envMap)
	return d
}

func (d *dio) RegisterBean(beanInstance ...interface{}) Dio {
	for _, bean := range beanInstance {
		d.RegisterNamedBean("", bean)
	}
	return d
}

func (d *dio) RegisterNamedBean(beanName string, beanInstance interface{}) Dio {
	if d.loaded {
		d.di.RegisterNamedBean(beanName, beanInstance)
	} else {
		d.providedBeans = append(d.providedBeans,
			bean{name: beanName,
				instance:   beanInstance,
				needMatch:  false,
				registered: true,
			})
	}
	return d
}

func (d *dio) Provide(prototype ...interface{}) Dio {
	for _, bean := range prototype {
		d.ProvideNamedBean("", bean)
	}
	return d
}

func (d *dio) ProvideNamedBean(beanName string, prototype interface{}) Dio {
	return d.ProvideNamedBeanOnProperty(beanName, prototype, "", "")
}

func (d *dio) ProvideMultiNamedBean(namedBeanMap map[string]interface{}) Dio {
	for name, bean := range namedBeanMap {
		d.ProvideNamedBean(name, bean)
	}
	return d
}

func (d *dio) provideBeanCondition(beanName string, prototype interface{}, property string, compareValue string, needMatch bool, caseSensitive bool) Dio {
	if d.loaded {
		panic("dio is already run")
	}
	d.providedBeans = append(d.providedBeans,
		bean{name: beanName,
			instance:        prototype,
			property:        property,
			compareValue:    compareValue,
			needMatch:       needMatch,
			caseInsensitive: !caseSensitive,
		})
	return d
}

func (d *dio) OnProperty(property string, compareValue string, caseSensitive bool, fn func(Dio)) Dio {
	if d.matchProperty(property, compareValue, true, !caseSensitive) {
		fn(d)
	}
	return d
}

func (d *dio) NotOnProperty(property string, compareValue string, caseSensitive bool, fn func(Dio)) Dio {
	if d.matchProperty(property, compareValue, false, !caseSensitive) {
		fn(d)
	}
	return d
}

func (d *dio) ProvideOnProperty(prototype interface{}, property string, compareValue string, caseSensitive ...bool) Dio {
	return d.ProvideNamedBeanOnProperty("", prototype, property, compareValue, caseSensitive...)
}

func (d *dio) ProvideMultiBeanOnProperty(beans []interface{}, property string, compareValue string, caseSensitive ...bool) Dio {
	for _, bean := range beans {
		d.ProvideOnProperty(bean, property, compareValue, caseSensitive...)
	}
	return d
}

func (d *dio) ProvideNamedBeanOnProperty(beanName string, prototype interface{}, property string, compareValue string, caseSensitive ...bool) Dio {
	return d.provideBeanCondition(beanName, prototype, property, compareValue, true, len(caseSensitive) > 0 && caseSensitive[0])
}

func (d *dio) ProvideMultiNamedBeanOnProperty(namedBeanMap map[string]interface{}, property string, compareValue string, caseSensitive ...bool) Dio {
	for name, bean := range namedBeanMap {
		d.ProvideNamedBeanOnProperty(name, bean, property, compareValue, caseSensitive...)
	}
	return d
}

func (d *dio) ProvideNotOnProperty(prototype interface{}, property string, compareValue string, caseSensitive ...bool) Dio {
	return d.ProvideNamedBeanNotOnProperty("", prototype, property, compareValue, caseSensitive...)
}

func (d *dio) ProvideMultiBeanNotOnProperty(beans []interface{}, property string, compareValue string, caseSensitive ...bool) Dio {
	for _, bean := range beans {
		d.ProvideNotOnProperty(bean, property, compareValue, caseSensitive...)
	}
	return d
}

func (d *dio) ProvideNamedBeanNotOnProperty(beanName string, prototype interface{}, property string, compareValue string, caseSensitive ...bool) Dio {
	return d.provideBeanCondition(beanName, prototype, property, compareValue, false, len(caseSensitive) > 0 && caseSensitive[0])
}

func (d *dio) ProvideMultiNamedBeanNotOnProperty(namedBeanMap map[string]interface{}, property string, compareValue string, caseSensitive ...bool) Dio {
	for name, bean := range namedBeanMap {
		d.ProvideNamedBeanNotOnProperty(name, bean, property, compareValue, caseSensitive...)
	}
	return d
}

func (d *dio) GetBean(beanName string) (bean interface{}, ok bool) {
	return d.di.GetBean(beanName)
}

func (d *dio) GetByType(beanType interface{}) (bean interface{}, ok bool) {
	return d.di.GetByType(beanType)
}

func (d *dio) Run(ctx context.Context) {
	if d.loaded {
		panic("dio is already run")
	}
	d.loaded = true

	// 配置日志组件
	systemLog := d.di.NewBean(system.Log{}).(*system.Log)
	dioLog := newDiLogger(ctx, systemLog)
	d.di.Log(dioLog)
	d.di.RegisterBean(systemLog)

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 配置bean
	for i := range d.providedBeans {
		beanDefinition := d.providedBeans[i]
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
	d.di.Serve(ctx)
}

func (d *dio) Use(plugins ...PluginConfig) Dio {
	for i := range plugins {
		plugins[i](d)
	}
	return d
}

func (d *dio) LoadDefaultConfig(configs embed.FS, filename string) Dio {
	f, err := configs.Open(filename)
	if err != nil {
		panic(err)
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	configMap := map[string]interface{}{}
	if err := yaml.Unmarshal(data, &configMap); err != nil {
		panic(err)
	}
	d.SetDefaultPropertyMap(configMap)
	return d
}

func (d *dio) LoadConfig(configs embed.FS, filename string) Dio {
	f, err := configs.Open(filename)
	if err != nil {
		panic(err)
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	configMap := map[string]interface{}{}
	if err := yaml.Unmarshal(data, &configMap); err != nil {
		panic(err)
	}
	d.SetPropertyMap(configMap)
	return d
}
