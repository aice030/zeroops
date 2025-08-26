package service

import (
	"context"
	"fmt"
	"mocks3/services/queue/internal/repository"
	"mocks3/shared/interfaces"
	"mocks3/shared/models"
	"mocks3/shared/observability"
	"time"
)

// QueueService 队列服务实现
type QueueService struct {
	repo   *repository.RedisQueueRepository
	logger *observability.Logger
}

// NewQueueService 创建队列服务
func NewQueueService(repo *repository.RedisQueueRepository, logger *observability.Logger) *QueueService {
	return &QueueService{
		repo:   repo,
		logger: logger,
	}
}

// DeleteTask相关方法

// EnqueueDeleteTask 入队删除任务
func (s *QueueService) EnqueueDeleteTask(ctx context.Context, task *models.DeleteTask) error {
	s.logger.Info(ctx, "Enqueuing delete task",
		observability.String("task_id", task.ID),
		observability.String("object_key", task.ObjectKey))

	// 生成任务ID（如果没有）
	if task.ID == "" {
		task.GenerateID()
	}

	// 设置默认状态
	if task.Status == "" {
		task.Status = models.TaskStatusPending
	}

	// 设置创建时间
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}

	err := s.repo.EnqueueDeleteTask(ctx, task)
	if err != nil {
		s.logger.Error(ctx, "Failed to enqueue delete task",
			observability.String("task_id", task.ID),
			observability.Error(err))
		return fmt.Errorf("enqueue delete task: %w", err)
	}

	s.logger.Info(ctx, "Delete task enqueued successfully",
		observability.String("task_id", task.ID))
	return nil
}

// DequeueDeleteTask 出队删除任务
func (s *QueueService) DequeueDeleteTask(ctx context.Context) (*models.DeleteTask, error) {
	// 使用5秒超时
	task, err := s.repo.DequeueDeleteTask(ctx, 5*time.Second)
	if err != nil {
		s.logger.Error(ctx, "Failed to dequeue delete task", observability.Error(err))
		return nil, fmt.Errorf("dequeue delete task: %w", err)
	}

	if task == nil {
		return nil, nil // 队列为空
	}

	s.logger.Info(ctx, "Delete task dequeued successfully",
		observability.String("task_id", task.ID),
		observability.String("object_key", task.ObjectKey))

	return task, nil
}

// UpdateDeleteTaskStatus 更新删除任务状态
func (s *QueueService) UpdateDeleteTaskStatus(ctx context.Context, taskID string, status models.TaskStatus, errorMsg string) error {
	s.logger.Info(ctx, "Updating delete task status",
		observability.String("task_id", taskID),
		observability.String("status", string(status)))

	err := s.repo.UpdateDeleteTaskStatus(ctx, taskID, status, errorMsg)
	if err != nil {
		s.logger.Error(ctx, "Failed to update delete task status",
			observability.String("task_id", taskID),
			observability.Error(err))
		return fmt.Errorf("update delete task status: %w", err)
	}

	s.logger.Info(ctx, "Delete task status updated successfully",
		observability.String("task_id", taskID),
		observability.String("status", string(status)))
	return nil
}

// SaveTask相关方法

// EnqueueSaveTask 入队保存任务
func (s *QueueService) EnqueueSaveTask(ctx context.Context, task *models.SaveTask) error {
	s.logger.Info(ctx, "Enqueuing save task",
		observability.String("task_id", task.ID),
		observability.String("object_key", task.ObjectKey))

	// 生成任务ID（如果没有）
	if task.ID == "" {
		task.GenerateID()
	}

	// 设置默认状态
	if task.Status == "" {
		task.Status = models.TaskStatusPending
	}

	// 设置创建时间
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}

	err := s.repo.EnqueueSaveTask(ctx, task)
	if err != nil {
		s.logger.Error(ctx, "Failed to enqueue save task",
			observability.String("task_id", task.ID),
			observability.Error(err))
		return fmt.Errorf("enqueue save task: %w", err)
	}

	s.logger.Info(ctx, "Save task enqueued successfully",
		observability.String("task_id", task.ID))
	return nil
}

// DequeueSaveTask 出队保存任务
func (s *QueueService) DequeueSaveTask(ctx context.Context) (*models.SaveTask, error) {
	// 使用5秒超时
	task, err := s.repo.DequeueSaveTask(ctx, 5*time.Second)
	if err != nil {
		s.logger.Error(ctx, "Failed to dequeue save task", observability.Error(err))
		return nil, fmt.Errorf("dequeue save task: %w", err)
	}

	if task == nil {
		return nil, nil // 队列为空
	}

	s.logger.Info(ctx, "Save task dequeued successfully",
		observability.String("task_id", task.ID),
		observability.String("object_key", task.ObjectKey))

	return task, nil
}

// UpdateSaveTaskStatus 更新保存任务状态
func (s *QueueService) UpdateSaveTaskStatus(ctx context.Context, taskID string, status models.TaskStatus, errorMsg string) error {
	s.logger.Info(ctx, "Updating save task status",
		observability.String("task_id", taskID),
		observability.String("status", string(status)))

	err := s.repo.UpdateSaveTaskStatus(ctx, taskID, status, errorMsg)
	if err != nil {
		s.logger.Error(ctx, "Failed to update save task status",
			observability.String("task_id", taskID),
			observability.Error(err))
		return fmt.Errorf("update save task status: %w", err)
	}

	s.logger.Info(ctx, "Save task status updated successfully",
		observability.String("task_id", taskID),
		observability.String("status", string(status)))
	return nil
}

// 统计和健康检查方法

// GetStats 获取队列统计信息
func (s *QueueService) GetStats(ctx context.Context) (map[string]any, error) {
	deleteQueueLen, err := s.repo.GetDeleteQueueLength(ctx)
	if err != nil {
		return nil, fmt.Errorf("get delete queue length: %w", err)
	}

	saveQueueLen, err := s.repo.GetSaveQueueLength(ctx)
	if err != nil {
		return nil, fmt.Errorf("get save queue length: %w", err)
	}

	stats := map[string]any{
		"service":             "queue-service",
		"delete_queue_length": deleteQueueLen,
		"save_queue_length":   saveQueueLen,
		"total_queue_length":  deleteQueueLen + saveQueueLen,
		"timestamp":           time.Now(),
	}

	s.logger.Info(ctx, "Queue stats retrieved",
		observability.Int64("delete_queue_length", deleteQueueLen),
		observability.Int64("save_queue_length", saveQueueLen))

	return stats, nil
}

// HealthCheck 健康检查
func (s *QueueService) HealthCheck(ctx context.Context) error {
	return s.repo.HealthCheck(ctx)
}

// 确保QueueService实现了interfaces.QueueService接口
var _ interfaces.QueueService = (*QueueService)(nil)
