package dio

import (
	"context"
	"fmt"
	"github.com/cheivin/di"
	"github.com/cheivin/dio-core"
	_ "github.com/kr/text"
	"gopkg.in/yaml.v2"
	"io/fs"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type dioContainer struct {
	log           core.Log
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

func New() core.Dio {
	container := &dioContainer{di: di.New(), providedBeans: []bean{}, loaded: false}
	container.di.Log(emptyLogger{})
	logName := "dio_app"
	if hostname, err := os.Hostname(); err == nil && hostname != "" {
		logName += "_" + hostname
	}
	container.SetDefaultProperty("log", map[string]interface{}{
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
	if needMatch {
		return match
	} else {
		return !match
	}
}

func (d *dioContainer) SetDefaultProperty(key string, value interface{}) core.Dio {
	d.di.SetDefaultProperty(key, value)
	return d
}

func (d *dioContainer) SetDefaultPropertyMap(properties map[string]interface{}) core.Dio {
	d.di.SetDefaultPropertyMap(properties)
	return d
}

func (d *dioContainer) SetProperty(key string, value interface{}) core.Dio {
	d.di.SetProperty(key, value)
	return d
}

func (d *dioContainer) SetPropertyMap(properties map[string]interface{}) core.Dio {
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

func (d *dioContainer) GetProperties(prefix string, destType interface{}) interface{} {
	return d.di.LoadProperties(prefix, destType)
}

func (d *dioContainer) AutoMigrateEnv() core.Dio {
	envMap := di.LoadEnvironment(strings.NewReplacer("_", "."), false)
	d.SetPropertyMap(envMap)
	return d
}

func (d *dioContainer) SetLogger(log core.Log) core.Dio {
	if d.loaded {
		panic("dioContainer is already run")
	}
	d.log = log
	d.di.RegisterBean(log)
	return d
}

func (d *dioContainer) Logger() core.Log {
	if !d.loaded {
		panic("dioContainer not run")
	}
	return d.log
}

func (d *dioContainer) RegisterBean(beanInstance ...interface{}) core.Dio {
	for _, bean := range beanInstance {
		d.RegisterNamedBean("", bean)
	}
	return d
}

func (d *dioContainer) RegisterNamedBean(beanName string, beanInstance interface{}) core.Dio {
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

func (d *dioContainer) Provide(prototype ...interface{}) core.Dio {
	for _, bean := range prototype {
		d.ProvideNamedBean("", bean)
	}
	return d
}

func (d *dioContainer) ProvideNamedBean(beanName string, prototype interface{}) core.Dio {
	return d.ProvideNamedBeanOnProperty(beanName, prototype, "", "")
}

func (d *dioContainer) ProvideMultiNamedBean(namedBeanMap map[string]interface{}) core.Dio {
	for name, bean := range namedBeanMap {
		d.ProvideNamedBean(name, bean)
	}
	return d
}

func (d *dioContainer) provideBeanCondition(beanName string, prototype interface{}, property string, compareValue string, needMatch bool, caseSensitive bool) core.Dio {
	if d.loaded {
		panic("dioContainer is already run")
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

func (d *dioContainer) ProvideOnProperty(prototype interface{}, property string, compareValue string, caseSensitive ...bool) core.Dio {
	return d.ProvideNamedBeanOnProperty("", prototype, property, compareValue, caseSensitive...)
}

func (d *dioContainer) ProvideMultiBeanOnProperty(beans []interface{}, property string, compareValue string, caseSensitive ...bool) core.Dio {
	for _, bean := range beans {
		d.ProvideOnProperty(bean, property, compareValue, caseSensitive...)
	}
	return d
}

func (d *dioContainer) ProvideNamedBeanOnProperty(beanName string, prototype interface{}, property string, compareValue string, caseSensitive ...bool) core.Dio {
	return d.provideBeanCondition(beanName, prototype, property, compareValue, true, len(caseSensitive) > 0 && caseSensitive[0])
}

func (d *dioContainer) ProvideMultiNamedBeanOnProperty(namedBeanMap map[string]interface{}, property string, compareValue string, caseSensitive ...bool) core.Dio {
	for name, bean := range namedBeanMap {
		d.ProvideNamedBeanOnProperty(name, bean, property, compareValue, caseSensitive...)
	}
	return d
}

func (d *dioContainer) ProvideNotOnProperty(prototype interface{}, property string, compareValue string, caseSensitive ...bool) core.Dio {
	return d.ProvideNamedBeanNotOnProperty("", prototype, property, compareValue, caseSensitive...)
}

func (d *dioContainer) ProvideMultiBeanNotOnProperty(beans []interface{}, property string, compareValue string, caseSensitive ...bool) core.Dio {
	for _, bean := range beans {
		d.ProvideNotOnProperty(bean, property, compareValue, caseSensitive...)
	}
	return d
}

func (d *dioContainer) ProvideNamedBeanNotOnProperty(beanName string, prototype interface{}, property string, compareValue string, caseSensitive ...bool) core.Dio {
	return d.provideBeanCondition(beanName, prototype, property, compareValue, false, len(caseSensitive) > 0 && caseSensitive[0])
}

func (d *dioContainer) ProvideMultiNamedBeanNotOnProperty(namedBeanMap map[string]interface{}, property string, compareValue string, caseSensitive ...bool) core.Dio {
	for name, bean := range namedBeanMap {
		d.ProvideNamedBeanNotOnProperty(name, bean, property, compareValue, caseSensitive...)
	}
	return d
}

func (d *dioContainer) GetBean(beanName string) (bean interface{}, ok bool) {
	return d.di.GetBean(beanName)
}

func (d *dioContainer) GetByType(beanType interface{}) (bean interface{}, ok bool) {
	return d.di.GetByType(beanType)
}

func (d *dioContainer) NewBean(beanType interface{}) (bean interface{}) {
	return d.di.NewBean(beanType)
}

func (d *dioContainer) NewBeanByName(beanName string) (bean interface{}) {
	return d.di.NewBeanByName(beanName)
}

func (d *dioContainer) Run(ctx context.Context, afterRunFns ...func(core.Dio)) {
	if d.loaded {
		panic("dioContainer is already run")
	}
	d.loaded = true

	// 配置日志组件
	if d.log == nil {
		property := d.GetProperties("log.", core.Property{}).(core.Property)
		if log, err := NewZapLogger(property); err != nil {
			panic(err)
		} else {
			d.log = log
		}
		systemLog := d.di.RegisterBean(d.log)
		d.di.RegisterBean(systemLog)
	}
	d.di.Log(newDiLogger(ctx, d.log))

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
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

	// 容器加载完成后执行的方法
	for i := range afterRunFns {
		afterRunFns[i](d)
	}

	d.di.Serve(ctx)
}

func (d *dioContainer) Use(plugins ...core.PluginConfig) core.Dio {
	for i := range plugins {
		plugins[i](d)
	}
	return d
}

func (d *dioContainer) LoadDefaultConfig(configs fs.FS, filename string) core.Dio {
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

func (d *dioContainer) LoadConfig(configs fs.FS, filename string) core.Dio {
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
