package main

import (
	"context"
	"fmt"
	"mocks3/services/queue/internal/handler"
	"mocks3/services/queue/internal/repository"
	"mocks3/services/queue/internal/service"
	"mocks3/services/queue/internal/worker"
	"mocks3/shared/client"
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
	config, err := service.LoadConfig("config/queue-config.yaml")
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

	logger.Info(ctx, "Starting Queue Service",
		observability.String("service", config.Service.Name),
		observability.String("host", config.Service.Host),
		observability.Int("port", config.Service.Port))

	// 3. 初始化Redis队列仓库
	repo, err := repository.NewRedisQueueRepository(config.GetRedisURL(), logger)
	if err != nil {
		logger.Error(ctx, "Failed to initialize Redis queue repository", observability.Error(err))
		os.Exit(1)
	}
	defer repo.Close()

	logger.Info(ctx, "Redis queue repository initialized")

	// 4. 初始化队列服务
	queueService := service.NewQueueService(repo, logger)

	// 5. 初始化外部服务客户端
	storageClient := client.NewStorageClient(
		"http://localhost:8082", // Storage Service地址
		30*time.Second,
		logger,
	)

	metadataClient := client.NewMetadataClient(
		"http://localhost:8081", // Metadata Service地址
		30*time.Second,
		logger,
	)

	// 6. 初始化任务处理器
	deleteProcessor := worker.NewStorageDeleteProcessor(storageClient, logger)
	saveProcessor := worker.NewStorageSaveProcessor(storageClient, metadataClient, logger)

	// 7. 初始化任务工作者
	taskWorker := worker.NewTaskWorker(
		queueService,
		deleteProcessor,
		saveProcessor,
		&config.Worker,
		logger,
	)

	// 8. 启动任务工作者
	taskWorker.Start()
	defer taskWorker.Stop()

	// 9. 初始化HTTP处理器
	queueHandler := handler.NewQueueHandler(queueService, logger)

	// 10. 设置Gin路由
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// 使用shared/observability中间件
	observability.SetupGinMiddlewares(router, config.Service.Name, httpMiddleware)

	// 设置业务路由
	queueHandler.SetupRoutes(router)

	// 11. 启动系统指标收集
	observability.StartSystemMetrics(ctx, collector, logger)

	// 12. 创建HTTP服务器
	addr := fmt.Sprintf("%s:%d", config.Service.Host, config.Service.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// 13. 启动服务器
	go func() {
		logger.Info(ctx, "HTTP server starting", observability.String("addr", addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(ctx, "HTTP server failed", observability.Error(err))
		}
	}()

	logger.Info(ctx, "Queue Service started successfully",
		observability.String("addr", addr))

	// 14. 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info(ctx, "Shutting down Queue Service...")

	// 关闭HTTP服务器
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error(ctx, "HTTP server shutdown failed", observability.Error(err))
	} else {
		logger.Info(ctx, "HTTP server stopped")
	}

	logger.Info(ctx, "Queue Service stopped")
}
