package mysql

import (
	"fmt"
	"github.com/cheivin/di"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/url"
	"time"
)

var opts []gorm.Option

func SetOptions(options ...gorm.Option) {
	opts = append(opts, options...)
}

type GormConfiguration struct {
	Username    string `value:"mysql.username"`
	Password    string `value:"mysql.password"`
	Host        string `value:"mysql.host"`
	Port        int    `value:"mysql.port"`
	Database    string `value:"mysql.database"`
	Parameters  string `value:"mysql.parameters"`
	MaxIdle     int    `value:"mysql.pool.max-idle"`
	MaxOpen     int    `value:"mysql.pool.max-open"`
	MaxLifeTime string `value:"mysql.pool.max-life-time"`
	MaxIdleTime string `value:"mysql.pool.max-idle-time"`
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

func (c *GormConfiguration) BeanConstruct(container *di.DI) {
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
	container.RegisterNamedBean("mysql", db)
}

// AfterPropertiesSet 注入完成时触发
func (c *GormConfiguration) AfterPropertiesSet() {
	db, _ := c.db.DB()
	if err := db.Ping(); err != nil {
		panic(err)
	}
	c.db.Logger = c.Logger
}
