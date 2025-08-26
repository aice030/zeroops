package interfaces

import (
	"context"
	"mocks3/shared/models"
)

// QueueService 队列服务接口
type QueueService interface {
	// 删除任务操作
	EnqueueDeleteTask(ctx context.Context, task *models.DeleteTask) error
	DequeueDeleteTask(ctx context.Context) (*models.DeleteTask, error)

	// 保存任务操作
	EnqueueSaveTask(ctx context.Context, task *models.SaveTask) error
	DequeueSaveTask(ctx context.Context) (*models.SaveTask, error)

	// 任务状态更新
	UpdateDeleteTaskStatus(ctx context.Context, taskID string, status models.TaskStatus, errorMsg string) error
	UpdateSaveTaskStatus(ctx context.Context, taskID string, status models.TaskStatus, errorMsg string) error

	// 状态查询
	GetStats(ctx context.Context) (map[string]any, error)

	// 健康检查
	HealthCheck(ctx context.Context) error
}

// DeleteTaskProcessor 删除任务处理器接口
type DeleteTaskProcessor interface {
	ProcessDeleteTask(ctx context.Context, task *models.DeleteTask) error
}

// SaveTaskProcessor 保存任务处理器接口
type SaveTaskProcessor interface {
	ProcessSaveTask(ctx context.Context, task *models.SaveTask) error
}
