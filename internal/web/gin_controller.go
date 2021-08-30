package web

import (
	"net/http"

	"github.com/cheivin/dio/system"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	Web *gin.Engine `aware:"web"`
	Log *system.Log `aware:""`
}

func (o *Controller) JsonView(fn func(*gin.Context) interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, fn(c))
	}
}

func (o *Controller) HtmlView(fn func(*gin.Context) (string, gin.H)) gin.HandlerFunc {
	return func(c *gin.Context) {
		view, data := fn(c)
		c.HTML(http.StatusOK, view, data)
	}
}

func (o *Controller) XmlView(fn func(*gin.Context) interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.XML(http.StatusOK, fn(c))
	}
}

func (o *Controller) JsonpView(fn func(*gin.Context) interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSONP(http.StatusOK, fn(c))
	}
}

func (o *Controller) StringView(fn func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.String(http.StatusOK, fn(c))
	}
}
