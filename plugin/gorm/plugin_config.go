package orm

import (
	"github.com/cheivin/dio"
	"gorm.io/gorm"
)

func Gorm(options ...gorm.Option) dio.PluginConfig {
	return func(d dio.Dio) {
		if !d.HasProperty("gorm") {
			d.SetDefaultProperty("gorm", map[string]interface{}{
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
		d.RegisterBean(&GormOptions{Options: options})
		d.Provide(GormConfiguration{})
		d.Provide(GormLogger{})
	}
}
