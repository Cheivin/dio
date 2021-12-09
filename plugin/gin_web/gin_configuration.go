package web

import (
	"context"
	"fmt"
	"github.com/cheivin/di"
	"github.com/cheivin/dio/system"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type ginContainer struct {
	Port   int         `value:"app.port"`
	Log    *system.Log `aware:""`
	router *gin.Engine
	server *http.Server
}

func (w *ginContainer) BeanName() string {
	return "ginWebContainer"
}

// BeanConstruct 初始化实例时，创建gin框架
func (w *ginContainer) BeanConstruct(container di.DI) {
	w.router = gin.New()
	w.router.RemoteIPHeaders = []string{"X-Forwarded-For", "X-Real-IP", "Proxy-Client-IP", "WL-Proxy-Client-IP", "HTTP_CLIENT_IP", "HTTP_X_FORWARDED_FOR"}
	// 注册gin到容器
	container.RegisterNamedBean("web", w.router)
}

// AfterPropertiesSet 注入完成时触发
func (w *ginContainer) AfterPropertiesSet(container di.DI) {
	w.server = &http.Server{
		Handler: w.router,
		Addr:    fmt.Sprintf(":%d", w.Port),
	}
	w.Log.Info(container.Context(), "Gin Web Container loaded")
	w.Log = w.Log.WithOptions(zap.WithCaller(false))
}

// Initialized DI加载完成后，启动服务
func (w *ginContainer) Initialized() {
	go func() {
		w.Log.Info(context.Background(), fmt.Sprintf("Gin Web Container starting at port: %d", w.Port))
		if err := w.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			w.Log.Error(context.Background(), "Gin Web Container fatal", "error", err)
			panic(err)
		}
	}()
}

func (w *ginContainer) Destroy() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := w.server.Shutdown(ctx); err != nil {
		w.Log.Error(ctx, "Server forced to shutdown", "error", err)
	} else {
		w.Log.Info(ctx, "Gin Web Container shutdown")
	}
}
