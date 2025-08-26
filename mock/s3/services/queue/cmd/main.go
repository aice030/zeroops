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
	"mocks3/shared/server"
	"os"
	"time"
)

// 全局变量用于在初始化和清理时共享
var (
	globalTaskWorker *worker.TaskWorker
	globalRepo       *repository.RedisQueueRepository
)

func main() {
	// 1. 加载配置
	config, err := service.LoadConfig("config/queue-config.yaml")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 2. 创建服务启动器
	bootstrap := server.NewServiceBootstrap("Queue Service", config)

	// 3. 设置自定义初始化逻辑
	bootstrap.WithCustomInit(func(ctx context.Context, logger *observability.Logger) error {
		// 初始化Redis队列仓库
		repo, err := repository.NewRedisQueueRepository(config.GetRedisURL(), logger)
		if err != nil {
			return fmt.Errorf("failed to initialize Redis queue repository: %w", err)
		}
		globalRepo = repo // 保存供清理使用

		logger.Info(ctx, "Redis queue repository initialized")

		// 初始化队列服务
		queueService := service.NewQueueService(repo, logger)

		// 初始化外部服务客户端
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

		// 初始化任务处理器
		deleteProcessor := worker.NewStorageDeleteProcessor(storageClient, logger)
		saveProcessor := worker.NewStorageSaveProcessor(storageClient, metadataClient, logger)

		// 初始化任务工作者
		taskWorker := worker.NewTaskWorker(
			queueService,
			deleteProcessor,
			saveProcessor,
			&config.Worker,
			logger,
		)
		globalTaskWorker = taskWorker // 保存供清理使用

		// 启动任务工作者
		taskWorker.Start()

		// 初始化HTTP处理器
		queueHandler := handler.NewQueueHandler(queueService, logger)

		// 设置处理器到启动器
		bootstrap.WithHandler(queueHandler)

		return nil
	})

	// 4. 设置自定义清理逻辑
	bootstrap.WithCustomCleanup(func(ctx context.Context, logger *observability.Logger) error {
		// 停止任务工作者
		if globalTaskWorker != nil {
			globalTaskWorker.Stop()
			logger.Info(ctx, "Task worker stopped")
		}

		// 关闭数据库连接
		if globalRepo != nil {
			globalRepo.Close()
			logger.Info(ctx, "Redis repository closed")
		}

		return nil
	})

	// 5. 启动服务
	if err := bootstrap.Start(); err != nil {
		fmt.Printf("Failed to start service: %v\n", err)
		os.Exit(1)
	}
}
