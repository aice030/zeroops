package interfaces

import (
	"context"
	"mocks3/shared/models"
)

// QueueService 队列服务接口（简化版，仅支持删除任务）
type QueueService interface {
	// 删除任务操作
	EnqueueDeleteTask(ctx context.Context, task *models.DeleteTask) error
	DequeueDeleteTask(ctx context.Context) (*models.DeleteTask, error)
	
	// 简单状态查询
	GetQueueLength(ctx context.Context) (int64, error)
	
	// 健康检查
	HealthCheck(ctx context.Context) error
}

// DeleteTaskProcessor 删除任务处理器接口
type DeleteTaskProcessor interface {
	ProcessDeleteTask(ctx context.Context, task *models.DeleteTask) error
}
