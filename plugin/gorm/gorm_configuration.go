package orm

import (
	"fmt"
	"github.com/cheivin/di"
	"github.com/cheivin/dio/plugin/gorm/dao"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/url"
	"time"
)

type GormOptions struct {
	Options []gorm.Option
}

func (c *GormOptions) BeanName() string {
	return "gormOptions"
}

type GormConfiguration struct {
	Username    string `value:"gorm.username"`
	Password    string `value:"gorm.password"`
	Host        string `value:"gorm.host"`
	Port        int    `value:"gorm.port"`
	Database    string `value:"gorm.database"`
	Parameters  string `value:"gorm.parameters"`
	MaxIdle     int    `value:"gorm.pool.max-idle"`
	MaxOpen     int    `value:"gorm.pool.max-open"`
	MaxLifeTime string `value:"gorm.pool.max-life-time"`
	MaxIdleTime string `value:"gorm.pool.max-idle-time"`
	db          *gorm.DB
	Logger      *GormLogger `aware:""`
}

func (c *GormConfiguration) BeanName() string {
	return "gormConfiguration"
}

func (c *GormConfiguration) parseParameters() {
	if c.Parameters == "" {
		return
	}
	_, err := url.ParseQuery(c.Parameters)
	if err != nil {
		panic(err)
	}
}

func (c *GormConfiguration) BeanConstruct(container di.DI) {
	var opts []gorm.Option
	if options, ok := container.GetByType(GormOptions{}); ok {
		opts = options.(*GormOptions).Options
	}

	c.parseParameters()

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", []interface{}{
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.Parameters,
	}...)
	// 配置db
	db, err := gorm.Open(mysql.Open(dsn), opts...)
	if err != nil {
		panic(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	if c.MaxLifeTime != "" {
		if maxLifetime, err := time.ParseDuration(c.MaxLifeTime); err != nil {
			panic(err)
		} else {
			sqlDB.SetConnMaxLifetime(maxLifetime)
		}
	}
	if c.MaxIdleTime != "" {
		if maxIdleTime, err := time.ParseDuration(c.MaxIdleTime); err != nil {
			panic(err)
		} else {
			sqlDB.SetConnMaxIdleTime(maxIdleTime)
		}
	}
	sqlDB.SetMaxIdleConns(c.MaxIdle)
	sqlDB.SetMaxOpenConns(c.MaxOpen)
	// 注册db
	c.db = db
	container.RegisterNamedBean("gorm", db)
	baseDao := dao.New(db)
	container.RegisterBean(baseDao)
}

// AfterPropertiesSet 注入完成时触发
func (c *GormConfiguration) AfterPropertiesSet() {
	db, _ := c.db.DB()
	if err := db.Ping(); err != nil {
		panic(err)
	}
	c.db.Logger = c.Logger
}
