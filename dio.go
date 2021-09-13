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
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type dio struct {
	providedBeans []bean
	loaded        bool
}
type bean struct {
	name         string
	instance     interface{}
	property     string
	compareValue string
	needMatch    bool
}

func (b bean) matchProperty() (match bool) {
	// 空值表示未设定条件
	if b.property == "" {
		return true
	}
	// 取出比较的属性值
	val := di.Property().Get(b.property)
	if val == nil {
		match = b.compareValue == ""
	} else {
		match = fmt.Sprintf("%v", val) == b.compareValue
	}
	if b.needMatch {
		return match
	} else {
		return !match
	}
}

var g *dio

const defaultTraceName = "X-Request-Id"

func init() {
	g = &dio{providedBeans: []bean{}, loaded: false}
	logName := "dio_app"
	if hostname, err := os.Hostname(); err == nil && hostname != "" {
		logName += "_" + hostname
	}
	di.SetDefaultProperty("log", map[string]interface{}{
		"name":       logName,
		"dir":        "./logs",
		"max-age":    30,
		"debug":      true,
		"std":        true,
		"file":       true,
		"trace-name": defaultTraceName,
	})
	di.Provide(system.Log{})
}

func (d *dio) SetDefaultProperty(key string, value interface{}) *dio {
	di.SetDefaultProperty(key, value)
	return d
}

func (d *dio) SetDefaultPropertyMap(properties map[string]interface{}) *dio {
	di.SetDefaultPropertyMap(properties)
	return d
}

func (d *dio) SetProperty(key string, value interface{}) *dio {
	di.SetProperty(key, value)
	return d
}

func (d *dio) SetPropertyMap(properties map[string]interface{}) *dio {
	di.SetPropertyMap(properties)
	return d
}

func (d *dio) HasProperty(property string) bool {
	return di.Property().Get(property) != nil
}

func (d *dio) GetPropertyString(property string) string {
	val := di.Property().Get(property)
	if val == nil {
		return ""
	} else {
		return fmt.Sprintf("%v", val)
	}
}

func (d *dio) AutoMigrateEnv() *dio {
	envMap := di.LoadEnvironment(strings.NewReplacer("_", "."), false)
	di.SetPropertyMap(envMap)
	return d
}

func (d *dio) RegisterBean(bean interface{}) *dio {
	di.RegisterBean(bean)
	return d
}

func (d *dio) RegisterNamedBean(name string, bean interface{}) *dio {
	di.RegisterNamedBean(name, bean)
	return d
}

func (d *dio) Provide(prototype interface{}) *dio {
	d.ProvideNamedBean("", prototype)
	return d
}

func (d *dio) ProvideNamedBean(beanName string, prototype interface{}) *dio {
	return d.ProvideNamedBeanOnProperty(beanName, prototype, "", "")
}

func (d *dio) ProvideOnProperty(prototype interface{}, property string, compareValue string) *dio {
	return d.ProvideNamedBeanOnProperty("", prototype, property, compareValue)
}

func (d *dio) ProvideNamedBeanOnProperty(beanName string, prototype interface{}, property string, compareValue string) *dio {
	if g.loaded {
		panic("dio is already run")
	}
	g.providedBeans = append(g.providedBeans,
		bean{name: beanName,
			instance:     prototype,
			property:     property,
			compareValue: compareValue,
			needMatch:    true,
		})
	return d
}

func (d *dio) ProvideNotOnProperty(prototype interface{}, property string, compareValue string) *dio {
	return d.ProvideNamedBeanNotOnProperty("", prototype, property, compareValue)
}

func (d *dio) ProvideNamedBeanNotOnProperty(beanName string, prototype interface{}, property string, compareValue string) *dio {
	if g.loaded {
		panic("dio is already run")
	}
	g.providedBeans = append(g.providedBeans,
		bean{name: beanName,
			instance:     prototype,
			property:     property,
			compareValue: compareValue,
			needMatch:    false,
		})
	return d
}

func (d *dio) GetBean(beanName string) (bean interface{}, ok bool) {
	return di.GetBean(beanName)
}

func (d *dio) Run(ctx context.Context) {
	if g.loaded {
		panic("dio is already run")
	}
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	for i := range g.providedBeans {
		beanDefinition := g.providedBeans[i]
		if beanDefinition.matchProperty() {
			di.ProvideNamedBean(beanDefinition.name, beanDefinition.instance)
		}
	}
	di.LoadAndServ(ctx)
}

func (d *dio) Web(useLogger, useCors bool) *dio {
	if !d.HasProperty("app.port") {
		di.SetDefaultPropertyMap(map[string]interface{}{
			"app.port": 8080,
		})
	}
	di.Provide(web.Container{})
	if useLogger {
		if !d.HasProperty("app.web.log") {
			di.SetDefaultProperty("app.web.log", map[string]interface{}{
				"skip-path":  "",
				"trace-name": defaultTraceName,
			})
		}
		di.Provide(middleware.WebLogger{})
	}
	di.Provide(middleware.WebRecover{})
	if useCors {
		if !d.HasProperty("app.web.cors") {
			di.SetDefaultProperty("app.web.cors", map[string]interface{}{
				"origin":            "",
				"method":            "",
				"header":            "",
				"allow-credentials": true,
				"expose-header":     "",
				"max-age":           43200,
			})
		}
		di.Provide(middleware.WebCors{})
	}
	di.Provide(system.Controller{})
	return d
}

func (d *dio) MySQL(options ...gorm.Option) *dio {
	mysql.SetOptions(options...)
	if !d.HasProperty("mysql") {
		di.SetDefaultProperty("mysql", map[string]interface{}{
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
	di.Provide(mysql.GormConfiguration{})
	di.Provide(mysql.GormLogger{})
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
	g.SetDefaultPropertyMap(configMap)
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
	g.SetPropertyMap(configMap)
	return g
}
