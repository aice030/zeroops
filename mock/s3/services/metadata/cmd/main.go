package main

import (
	"context"
	"fmt"
	"mocks3/services/metadata/internal/handler"
	"mocks3/services/metadata/internal/repository"
	"mocks3/services/metadata/internal/service"
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
	config, err := service.LoadConfig("config/metadata-config.yaml")
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

	logger.Info(ctx, "Starting Metadata Service",
		observability.String("service", config.Service.Name),
		observability.String("host", config.Service.Host),
		observability.Int("port", config.Service.Port))

	// 3. 初始化数据库仓库
	repo, err := repository.NewSimplePostgreSQLRepository(config.GetDSN())
	if err != nil {
		logger.Error(ctx, "Failed to initialize repository", observability.Error(err))
		os.Exit(1)
	}
	defer repo.Close()

	logger.Info(ctx, "Database connection established")

	// 4. 初始化业务服务
	metadataService := service.NewMetadataService(repo, logger)

	// 5. 初始化HTTP处理器
	metadataHandler := handler.NewMetadataHandler(metadataService, logger)

	// 6. 设置Gin路由
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// 使用shared/observability中间件
	observability.SetupGinMiddlewares(router, config.Service.Name, httpMiddleware)

	// 设置业务路由
	metadataHandler.SetupRoutes(router)

	// 7. 启动系统指标收集
	observability.StartSystemMetrics(ctx, collector, logger)

	// 8. 创建HTTP服务器
	addr := fmt.Sprintf("%s:%d", config.Service.Host, config.Service.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// 9. 启动服务器
	go func() {
		logger.Info(ctx, "HTTP server starting", observability.String("addr", addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(ctx, "HTTP server failed", observability.Error(err))
		}
	}()

	logger.Info(ctx, "Metadata Service started successfully",
		observability.String("addr", addr))

	// 10. 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info(ctx, "Shutting down Metadata Service...")

	// 关闭HTTP服务器
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error(ctx, "HTTP server shutdown failed", observability.Error(err))
	} else {
		logger.Info(ctx, "HTTP server stopped")
	}

	logger.Info(ctx, "Metadata Service stopped")
}
