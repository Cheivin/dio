package system

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Controller struct {
	Web *gin.Engine `aware:"web"`
	Log *Log        `aware:""`
}

func (o *Controller) JsonView(fn func(*gin.Context) interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		data := fn(c)
		if !c.IsAborted() {
			c.JSON(http.StatusOK, data)
		}
	}
}

func (o *Controller) HtmlView(fn func(*gin.Context) (string, gin.H)) gin.HandlerFunc {
	return func(c *gin.Context) {
		view, data := fn(c)
		if !c.IsAborted() {
			c.HTML(http.StatusOK, view, data)
		}
	}
}

func (o *Controller) XmlView(fn func(*gin.Context) interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		data := fn(c)
		if !c.IsAborted() {
			c.XML(http.StatusOK, data)
		}
	}
}

func (o *Controller) JsonpView(fn func(*gin.Context) interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		data := fn(c)
		if !c.IsAborted() {
			c.JSONP(http.StatusOK, data)
		}
	}
}

func (o *Controller) StringView(fn func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		data := fn(c)
		if !c.IsAborted() {
			c.String(http.StatusOK, data)
		}
	}
}
