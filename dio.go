package dio

import (
	"context"
	"github.com/cheivin/di"
	"github.com/cheivin/dio/internal/mysql"
	"github.com/cheivin/dio/internal/web"
	"github.com/cheivin/dio/middleware"
	"github.com/cheivin/dio/system"
	"gorm.io/gorm"
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
	name     string
	instance interface{}
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
	d.ProvideWithBeanName("", prototype)
	return d
}

func (d *dio) ProvideWithBeanName(beanName string, prototype interface{}) *dio {
	if g.loaded {
		panic("dio is already run")
	}
	g.providedBeans = append(g.providedBeans, bean{name: beanName, instance: prototype})
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
		di.ProvideWithBeanName(beanDefinition.name, beanDefinition.instance)
	}
	di.LoadAndServ(ctx)
}

func (d *dio) Web(useLogger, useCors bool) *dio {
	di.SetDefaultPropertyMap(map[string]interface{}{
		"app.port": 8080,
		"app.env":  "dev",
	})
	di.Provide(web.Container{})
	di.Provide(middleware.WebRecover{})
	if useLogger {
		di.SetDefaultProperty("app.web.log", map[string]interface{}{
			"skip-path":  "",
			"trace-name": defaultTraceName,
		})
		di.Provide(middleware.WebLogger{})
	}
	if useCors {
		di.SetDefaultProperty("app.web.cors", map[string]interface{}{
			"origin":            "",
			"method":            "",
			"header":            "",
			"allow-credentials": true,
			"expose-header":     "",
			"max-age":           43200,
		})
		di.Provide(middleware.WebCors{})
	}

	return d
}

func (d *dio) MySQL(options ...gorm.Option) *dio {
	mysql.SetOptions(options...)
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
	di.Provide(mysql.GormConfiguration{})
	di.Provide(mysql.GormLogger{})
	return d
}
