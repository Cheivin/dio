package web

import (
	"github.com/cheivin/dio"
	"github.com/cheivin/dio/plugin/gin_web/middleware"
)

const defaultTraceName = dio.DefaultTraceName

func GinWeb(useLogger, useCors bool) dio.PluginConfig {
	return func(d dio.Dio) {
		if !d.HasProperty("app.port") {
			d.SetDefaultPropertyMap(map[string]interface{}{
				"app.port": 8080,
			})
		}
		d.Provide(ginContainer{})
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
		d.Provide(Controller{})
	}
}
