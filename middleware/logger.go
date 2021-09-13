package middleware

import (
	"github.com/cheivin/dio/system"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"strings"
	"time"
)

// WebLogger 日志
type WebLogger struct {
	Web         *gin.Engine `aware:"web"`
	Log         *system.Log `aware:"log"`
	Skips       string      `value:"app.web.log.skip-path"`
	TraceName   string      `value:"app.web.log.trace-name"`
	skip        map[string]struct{}
	idGenerator func(c *gin.Context) string
}

func (w *WebLogger) BeanConstruct() {
	skipPaths := strings.Split(w.Skips, ",")
	w.skip = make(map[string]struct{}, len(skipPaths))
	for _, path := range skipPaths {
		if path != "" {
			w.skip[path] = struct{}{}
		}
	}
	w.idGenerator = func(_ *gin.Context) string {
		return uuid.NewV4().String()
	}
}

func (w *WebLogger) SetIdGenerator(idGenerator func(c *gin.Context) string) {
	w.idGenerator = idGenerator
}

func (w *WebLogger) AfterPropertiesSet() {
	w.Web.Use(w.log)
}

func (w *WebLogger) log(c *gin.Context) {
	defer func() {
		// 此处recover用于处理顶层log中间件写出日志panic
		if r := recover(); r != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
	}()
	// 开始时间
	start := time.Now()
	path := c.Request.URL.Path
	raw := c.Request.URL.RawQuery

	// 设置id
	if w.TraceName != "" {
		reqId := c.GetHeader(w.TraceName)

		if reqId == "" {
			reqId = w.idGenerator(c)
			c.Header(w.TraceName, reqId)
		}
		c.Set(w.Log.TraceName, reqId)
	}
	// 处理请求
	c.Next()

	// 判断是否过滤路径
	for skipPath := range w.skip {
		if strings.HasPrefix(path, skipPath) {
			return
		}
	}

	// 记录日志
	timeStamp := time.Now()
	if raw != "" {
		path = path + "?" + raw
	}
	params := []interface{}{
		"TimeStamp", timeStamp,
		"Cost", timeStamp.Sub(start).String(),
		"ClientIP", c.ClientIP(),
		"Method", c.Request.Method,
		"StatusCode", c.Writer.Status(),
		"Path", path,
		"BodySize", c.Writer.Size(),
	}
	errMsg := c.Errors.ByType(gin.ErrorTypePrivate).String()
	if errMsg != "" {
		params = append(params, "ErrorMessage", errMsg)
		w.Log.Error(c, "gin-http", params...)
	} else {
		w.Log.Info(c, "gin-http", params...)
	}
}
