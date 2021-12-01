package web

import (
	"context"
	"fmt"
	"github.com/cheivin/di"
	"github.com/cheivin/dio/system"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type Container struct {
	Port   int         `value:"app.port"`
	Log    *system.Log `aware:"log"`
	router *gin.Engine
	server *http.Server
}

func (w *Container) BeanName() string {
	return "ginWebContainer"
}

// BeanConstruct 初始化实例时，创建gin框架
func (w *Container) BeanConstruct(container di.DI) {
	w.router = gin.New()
	w.router.RemoteIPHeaders = []string{"X-Forwarded-For", "X-Real-IP", "Proxy-Client-IP", "WL-Proxy-Client-IP", "HTTP_CLIENT_IP", "HTTP_X_FORWARDED_FOR"}
	// 注册gin到容器
	container.RegisterNamedBean("web", w.router)
}

// AfterPropertiesSet 注入完成时触发
func (w *Container) AfterPropertiesSet() {
	w.server = &http.Server{
		Handler: w.router,
		Addr:    fmt.Sprintf(":%d", w.Port),
	}
}

// Initialized DI加载完成后，启动服务
func (w *Container) Initialized() {
	go func() {
		w.Log.Info(context.Background(), fmt.Sprintf("Container starting at port: %d", w.Port))
		if err := w.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			w.Log.Error(context.Background(), "Container fatal", "error", err)
			panic(err)
		}
	}()
}

func (w *Container) Destroy() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := w.server.Shutdown(ctx); err != nil {
		w.Log.Error(ctx, "Server forced to shutdown", "error", err)
	} else {
		w.Log.Info(ctx, "Container shutdown")
	}
}
