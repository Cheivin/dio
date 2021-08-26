package middleware

import (
	code "github.com/cheivin/dio/errors"
	"github.com/cheivin/dio/system"
	"github.com/gin-gonic/gin"
	"net/http"
)

// WebRecover 全局异常
type WebRecover struct {
	Web        *gin.Engine `aware:"web"`
	Log        *system.Log `aware:"log"`
	responseFn func(c *gin.Context, err code.Error)
}

func (w *WebRecover) BeanConstruct() {
	w.responseFn = func(c *gin.Context, err code.Error) {
		c.String(err.Code, err.Error())
	}
}

func (w *WebRecover) SetErrorHandler(fn func(c *gin.Context, err code.Error)) {
	w.responseFn = fn
}

func (w *WebRecover) AfterPropertiesSet() {
	w.Web.NoRoute(w.noRoute)
	w.Web.NoMethod(w.noMethod)
	w.Web.Use(w.recover)
}

func (w *WebRecover) noRoute(c *gin.Context) {
	w.responseFn(c, code.NoRoute)
}

func (w *WebRecover) noMethod(c *gin.Context) {
	w.responseFn(c, code.NoMethod)
}

func (w *WebRecover) recover(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case code.Error:
				err := r.(code.Error)
				_ = c.Error(&err)
				w.responseFn(c, err)
			case error:
				e := r.(error)
				err := code.Err(http.StatusInternalServerError, e.Error(), e)
				_ = c.Error(&err)
				w.responseFn(c, err)
			case string:
				err := code.ErrMsg(http.StatusInternalServerError, r.(string))
				_ = c.Error(&err)
				w.responseFn(c, err)
			default:
				w.Log.Error(c, "Web server panic", "panic", r)
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			c.Abort()
		}
	}()
	//加载完 defer recover，继续后续接口调用
	c.Next()
}
