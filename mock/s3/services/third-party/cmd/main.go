package main

import (
	"context"
	"fmt"
	"mocks3/services/third-party/internal/handler"
	"mocks3/services/third-party/internal/service"
	"mocks3/shared/observability"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 加载配置
	config, err := service.LoadConfig("config/third-party-config.yaml")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 2. 初始化可观测性组件
	providers, collector, httpMiddleware, err := observability.Setup(config.Service.Name, "observability.yaml")
	if err != nil {
		fmt.Printf("Failed to setup observability: %v\n", err)
		os.Exit(1)
	}
	defer observability.Shutdown(context.Background(), providers)

	logger := observability.GetLogger(providers)
	ctx := context.Background()

	logger.Info(ctx, "Starting Third-Party Service",
		observability.String("service", config.Service.Name),
		observability.String("host", config.Service.Host),
		observability.Int("port", config.Service.Port),
		observability.String("mock_enabled", fmt.Sprintf("%v", config.Mock.Enabled)))

	// 3. 初始化第三方服务
	thirdPartyService := service.NewThirdPartyService(config, logger)

	// 4. 初始化HTTP处理器
	thirdPartyHandler := handler.NewThirdPartyHandler(thirdPartyService, logger)

	// 5. 设置Gin路由
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// 使用shared/observability中间件
	observability.SetupGinMiddlewares(router, config.Service.Name, httpMiddleware)

	// 设置业务路由
	thirdPartyHandler.SetupRoutes(router)

	// 6. 启动系统指标收集
	observability.StartSystemMetrics(ctx, collector, logger)

	// 7. 创建HTTP服务器
	addr := fmt.Sprintf("%s:%d", config.Service.Host, config.Service.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// 8. 启动服务器
	go func() {
		logger.Info(ctx, "HTTP server starting", observability.String("addr", addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(ctx, "HTTP server failed", observability.Error(err))
		}
	}()

	logger.Info(ctx, "Third-Party Service started successfully",
		observability.String("addr", addr),
		observability.Int("data_sources", len(config.DataSources)),
		observability.String("mock_mode", fmt.Sprintf("%v", config.Mock.Enabled)))

	// 9. 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info(ctx, "Shutting down Third-Party Service...")

	// 关闭HTTP服务器
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error(ctx, "HTTP server shutdown failed", observability.Error(err))
	} else {
		logger.Info(ctx, "HTTP server stopped")
	}

	logger.Info(ctx, "Third-Party Service stopped")
}