package mysql

import (
	"context"
	"errors"
	"fmt"
	"github.com/cheivin/dio/system"
	"gorm.io/gorm/logger"
	"time"
)

type GormLogger struct {
	Level                     int         `value:"mysql.Log.level"`
	SlowThreshold             string      `value:"mysql.Log.slow-log"`
	IgnoreRecordNotFoundError bool        `value:"mysql.Log.ignore-notfound"`
	Log                       *system.Log `aware:"log"`

	slowThreshold time.Duration
	level         logger.LogLevel
}

func (l *GormLogger) BeanName() string {
	return "gormLogger"
}

func (l *GormLogger) BeanConstruct() {
	// 日志配置
	if l.Level <= 0 || l.Level > 4 {
		l.level = logger.Info
	} else {
		l.level = logger.LogLevel(l.Level)
	}
	if l.SlowThreshold != "" {
		if slowThreshold, err := time.ParseDuration(l.SlowThreshold); err != nil {
			panic(err)
		} else {
			l.slowThreshold = slowThreshold
		}
	}
}

// AfterPropertiesSet 注入完成时触发
func (l *GormLogger) AfterPropertiesSet() {
	l.Log = l.Log.Skip(4)
}

// LogMode log mode
func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return &GormLogger{
		Level:                     int(level),
		SlowThreshold:             l.SlowThreshold,
		IgnoreRecordNotFoundError: l.IgnoreRecordNotFoundError,
		slowThreshold:             l.slowThreshold,
		level:                     level,
		Log:                       l.Log,
	}
}

// Info print info
func (l GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= logger.Info {
		l.Log.Info(ctx, fmt.Sprintf("%s\n "+msg, data...))
	}
}

// Warn print warn messages
func (l GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= logger.Warn {
		l.Log.Warn(ctx, fmt.Sprintf("%s\n "+msg, data...))
	}
}

// Error print error messages
func (l GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= logger.Error {
		l.Log.Error(ctx, fmt.Sprintf("%s\n "+msg, data...))
	}
}

// Trace print sql message
func (l GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.level > logger.Silent {
		elapsed := time.Since(begin)
		switch {
		case err != nil && l.level >= logger.Error && (!errors.Is(err, logger.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
			sql, rows := fc()
			if rows == -1 {
				l.Log.Error(ctx, err.Error(), "Cost", float64(elapsed.Nanoseconds())/1e6, "SQL", sql)
			} else {
				l.Log.Error(ctx, err.Error(), "Cost", float64(elapsed.Nanoseconds())/1e6, "Rows", rows, "SQL", sql)
			}
		case l.slowThreshold != 0 && elapsed > l.slowThreshold && l.level >= logger.Warn:
			sql, rows := fc()
			slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
			if rows == -1 {
				l.Log.Warn(ctx, slowLog, "Cost", float64(elapsed.Nanoseconds())/1e6, "SQL", sql)
			} else {
				l.Log.Warn(ctx, slowLog, "Cost", float64(elapsed.Nanoseconds())/1e6, "Rows", rows, "SQL", sql)
			}
		case l.level == logger.Info:
			sql, rows := fc()
			if rows == -1 {
				l.Log.Debug(ctx, "", "Cost", float64(elapsed.Nanoseconds())/1e6, "SQL", sql)
			} else {
				l.Log.Debug(ctx, "", "Cost", float64(elapsed.Nanoseconds())/1e6, "Rows", rows, "SQL", sql)
			}
		}
	}
}
