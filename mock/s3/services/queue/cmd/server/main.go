package main

import (
	"context"
	"fmt"
	"log"
	"mocks3/services/queue/internal/config"
	"mocks3/services/queue/internal/handler"
	"mocks3/services/queue/internal/repository"
	"mocks3/services/queue/internal/service"
	"mocks3/shared/middleware"
	logger "mocks3/shared/observability/log"
	"mocks3/shared/observability/metric"
	"mocks3/shared/observability/trace"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化日志器
	logger := logger.NewLogger("queue-service", logger.LevelInfo)

	// 初始化追踪器
	tracerProvider, err := trace.NewDefaultTracerProvider("queue-service")
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	defer tracerProvider.Shutdown(context.Background())

	// 初始化指标收集器
	metricCollector, err := metric.NewDefaultCollector("queue-service")
	if err != nil {
		log.Fatalf("Failed to initialize metrics: %v", err)
	}
	defer metricCollector.Shutdown(context.Background())

	// 初始化 Consul 配置
	consulAddr := "consul:8500"
	err = middleware.InitializeServiceConfig(consulAddr, "queue-service", 8083)
	if err != nil {
		log.Printf("Failed to initialize consul config: %v", err)
	}

	// 初始化Consul管理器
	consulManager, err := middleware.NewDefaultConsulManager("queue-service")
	if err != nil {
		log.Fatalf("Failed to initialize consul: %v", err)
	}

	// 初始化Redis仓库（使用默认配置）
	redisConfig := &config.RedisConfig{
		Host:     "redis",
		Port:     6379,
		Password: "",
		DB:       0,
	}
	queueConfig := &config.QueueConfig{
		StreamName:    "mocks3:tasks",
		ConsumerGroup: "queue-workers",
		MaxWorkers:    3,
	}
	redisRepo, err := repository.NewRedisRepository(redisConfig, queueConfig)
	if err != nil {
		log.Fatalf("Failed to initialize Redis repository: %v", err)
	}

	// 初始化服务
	queueService := service.NewQueueService(redisRepo, logger)

	// 初始化处理器
	queueHandler := handler.NewQueueHandler(queueService, logger)

	// 注册服务到Consul（从 Consul KV 加载配置）
	ctx := context.Background()
	consulConfig, err := middleware.LoadServiceConfigFromConsul(consulAddr, "queue-service")
	if err != nil {
		log.Fatalf("Failed to load consul config: %v", err)
	}

	err = consulManager.RegisterService(ctx, consulConfig)
	if err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}
	defer consulManager.DeregisterService(ctx)

	// 启动默认工作节点
	for i := 1; i <= 3; i++ {
		workerID := fmt.Sprintf("worker-%d", i)
		if err := queueService.StartWorker(ctx, workerID); err != nil {
			logger.Error("Failed to start worker", "worker_id", workerID, "error", err)
		} else {
			logger.Info("Started worker", "worker_id", workerID)
		}
	}

	// 设置Gin模式（默认开发模式）

	// 创建路由器
	router := gin.New()

	// 添加中间件
	router.Use(gin.Logger())
	router.Use(middleware.GinRecoveryMiddleware(middleware.DefaultRecoveryConfig()))
	router.Use(trace.GinMiddleware("queue-service"))

	// 添加指标中间件
	metricsMiddleware := metric.NewDefaultMiddlewareConfig(metricCollector)
	router.Use(metricsMiddleware.GinMiddleware())

	// 设置路由
	queueHandler.RegisterRoutes(router)

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		if err := queueService.HealthCheck(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "unhealthy",
				"service": "queue-service",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "queue-service",
			"version":   "1.0.0",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// 创建HTTP服务器（使用 Consul 配置的端口）
	addr := fmt.Sprintf(":%d", consulConfig.ServicePort)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动服务器
	go func() {
		logger.Info("Starting queue service", "address", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down queue service...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 停止队列服务
	if err := queueService.Stop(); err != nil {
		logger.Error("Failed to stop queue service", "error", err)
	}

	// 关闭HTTP服务器
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Queue service stopped")
}
