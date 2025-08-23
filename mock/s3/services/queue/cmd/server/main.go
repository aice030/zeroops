package main

import (
	"context"
	"fmt"
	"mocks3/services/queue/internal/config"
	"mocks3/services/queue/internal/handler"
	"mocks3/services/queue/internal/repository"
	"mocks3/services/queue/internal/service"
	"mocks3/shared/bootstrap"
)

// QueueServiceInitializer 实现服务初始化接口
type QueueServiceInitializer struct {
	queueService *service.QueueService
}

// Initialize 初始化队列服务的特定组件
func (q *QueueServiceInitializer) Initialize(bootstrap *bootstrap.ServiceBootstrap) error {
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
		return err
	}

	// 初始化服务
	q.queueService = service.NewQueueService(redisRepo, bootstrap.GetLogger())

	// 初始化处理器
	queueHandler := handler.NewQueueHandler(q.queueService, bootstrap.GetLogger())

	// 注册路由
	queueHandler.RegisterRoutes(bootstrap.GetRouter())

	// 启动默认工作节点
	ctx := context.Background()
	for i := 1; i <= 3; i++ {
		workerID := fmt.Sprintf("worker-%d", i)
		if err := q.queueService.StartWorker(ctx, workerID); err != nil {
			bootstrap.GetLogger().Error("Failed to start worker", "worker_id", workerID, "error", err)
		} else {
			bootstrap.GetLogger().Info("Started worker", "worker_id", workerID)
		}
	}

	return nil
}

func main() {
	// 创建服务配置
	serviceConfig := bootstrap.ServiceConfig{
		ServiceName: "queue-service",
		ServicePort: 8083,
		Version:     "1.0.0",
		Environment: "development",
		LogLevel:    "info",
		ConsulAddr:  "consul:8500",
	}

	// 创建初始化器
	initializer := &QueueServiceInitializer{}

	// 运行服务
	bootstrap.RunService(serviceConfig, initializer)
}
