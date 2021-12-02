package dio

import (
	"github.com/cheivin/dio/internal/mysql"
	"github.com/cheivin/dio/internal/web"
	"github.com/cheivin/dio/middleware"
	"github.com/cheivin/dio/system"
	"gorm.io/gorm"
)

type PluginConfig func(d *dio)

func Web(useLogger, useCors bool) PluginConfig {
	return func(d *dio) {
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
	}
}

func MySQL(options ...gorm.Option) PluginConfig {
	return func(d *dio) {
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
	}
}
