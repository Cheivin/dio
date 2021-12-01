package dio

import (
	"context"
	"embed"
	"fmt"
	"github.com/cheivin/di"
	"github.com/cheivin/dio/internal/mysql"
	"github.com/cheivin/dio/internal/web"
	"github.com/cheivin/dio/middleware"
	"github.com/cheivin/dio/system"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
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
	// 空值表示未设定条件
	if b.property == "" {
		return true
	}
	// 取出比较的属性值
	val := d.di.Property().Get(b.property)
	if val == nil {
		match = b.compareValue == ""
	} else {
		compareValue := fmt.Sprintf("%v", val)
		if b.caseInsensitive {
			match = strings.EqualFold(compareValue, b.compareValue)
		} else {
			match = compareValue == b.compareValue
		}
	}
	if b.needMatch {
		return match
	} else {
		return !match
	}
}

const defaultTraceName = "X-Request-Id"

func (d *dio) SetDefaultProperty(key string, value interface{}) *dio {
	d.di.SetDefaultProperty(key, value)
	return d
}

func (d *dio) SetDefaultPropertyMap(properties map[string]interface{}) *dio {
	d.di.SetDefaultPropertyMap(properties)
	return d
}

func (d *dio) SetProperty(key string, value interface{}) *dio {
	d.di.SetProperty(key, value)
	return d
}

func (d *dio) SetPropertyMap(properties map[string]interface{}) *dio {
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

func (d *dio) AutoMigrateEnv() *dio {
	envMap := di.LoadEnvironment(strings.NewReplacer("_", "."), false)
	d.SetPropertyMap(envMap)
	return d
}

func (d *dio) RegisterBean(beanInstance interface{}) *dio {
	d.RegisterNamedBean("", beanInstance)
	return d
}

func (d *dio) RegisterNamedBean(beanName string, beanInstance interface{}) *dio {
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

func (d *dio) Provide(prototype interface{}) *dio {
	d.ProvideNamedBean("", prototype)
	return d
}

func (d *dio) ProvideNamedBean(beanName string, prototype interface{}) *dio {
	return d.ProvideNamedBeanOnProperty(beanName, prototype, "", "")
}

func (d *dio) ProvideOnProperty(prototype interface{}, property string, compareValue string, caseInsensitive ...bool) *dio {
	return d.ProvideNamedBeanOnProperty("", prototype, property, compareValue, caseInsensitive...)
}

func (d *dio) ProvideNamedBeanOnProperty(beanName string, prototype interface{}, property string, compareValue string, caseInsensitive ...bool) *dio {
	if d.loaded {
		panic("dio is already run")
	}
	d.providedBeans = append(d.providedBeans,
		bean{name: beanName,
			instance:        prototype,
			property:        property,
			compareValue:    compareValue,
			needMatch:       true,
			caseInsensitive: len(caseInsensitive) > 0 && caseInsensitive[0] == true,
		})
	return d
}

func (d *dio) ProvideNotOnProperty(prototype interface{}, property string, compareValue string, caseInsensitive ...bool) *dio {
	return d.ProvideNamedBeanNotOnProperty("", prototype, property, compareValue, caseInsensitive...)
}

func (d *dio) ProvideNamedBeanNotOnProperty(beanName string, prototype interface{}, property string, compareValue string, caseInsensitive ...bool) *dio {
	if d.loaded {
		panic("dio is already run")
	}
	d.providedBeans = append(d.providedBeans,
		bean{name: beanName,
			instance:        prototype,
			property:        property,
			compareValue:    compareValue,
			needMatch:       false,
			caseInsensitive: len(caseInsensitive) > 0 && caseInsensitive[0] == true,
		})
	return d
}

func (d *dio) GetBean(beanName string) (bean interface{}, ok bool) {
	return d.di.GetBean(beanName)
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

func (d *dio) Web(useLogger, useCors bool) *dio {
	if !d.HasProperty("app.port") {
		d.SetDefaultPropertyMap(map[string]interface{}{
			"app.port": 8080,
		})
	}
	d.Provide(web.Container{})
	if useLogger {
		if !d.HasProperty("app.web.log") {
			d.SetDefaultProperty("app.web.log", map[string]interface{}{
				"skip-path":  "",
				"trace-name": defaultTraceName,
			})
		}
		d.Provide(middleware.WebLogger{})
	}
	d.Provide(middleware.WebRecover{})
	if useCors {
		if !d.HasProperty("app.web.cors") {
			d.SetDefaultProperty("app.web.cors", map[string]interface{}{
				"origin":            "",
				"method":            "",
				"header":            "",
				"allow-credentials": true,
				"expose-header":     "",
				"max-age":           43200,
			})
		}
		d.Provide(middleware.WebCors{})
	}
	d.Provide(system.Controller{})
	return d
}

func (d *dio) MySQL(options ...gorm.Option) *dio {
	mysql.SetOptions(options...)
	if !d.HasProperty("mysql") {
		d.SetDefaultProperty("mysql", map[string]interface{}{
			"username": "root",
			"password": "root",
			"host":     "localhost",
			"port":     3306,
			"pool": map[string]interface{}{
				"max-idle": 0,
				"max-open": 0,
			},
			"log.level": 4,
		})
	}
	d.Provide(mysql.GormConfiguration{})
	d.Provide(mysql.GormLogger{})
	return d
}

func (d *dio) LoadDefaultConfig(configs embed.FS, filename string) *dio {
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
	return g
}

func (d *dio) LoadConfig(configs embed.FS, filename string) *dio {
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
	return g
}
