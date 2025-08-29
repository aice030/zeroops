package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"mocks3/shared/models"
	"mocks3/shared/observability"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisQueueRepository Redis队列仓库实现
type RedisQueueRepository struct {
	client *redis.Client
	logger *observability.Logger
}

// NewRedisQueueRepository 创建Redis队列仓库
func NewRedisQueueRepository(redisURL string, logger *observability.Logger) (*RedisQueueRepository, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("connect to redis: %w", err)
	}

	logger.Info(context.Background(), "Redis queue repository connected successfully",
		observability.String("redis_url", redisURL))

	return &RedisQueueRepository{
		client: client,
		logger: logger,
	}, nil
}

// DeleteTask队列操作

// EnqueueDeleteTask 入队删除任务
func (r *RedisQueueRepository) EnqueueDeleteTask(ctx context.Context, task *models.DeleteTask) error {
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("marshal delete task: %w", err)
	}

	// 使用Redis List作为队列，LPUSH入队
	if err := r.client.LPush(ctx, "queue:delete_tasks", data).Err(); err != nil {
		return fmt.Errorf("enqueue delete task: %w", err)
	}

	// 同时存储到哈希表中用于状态查询和更新
	taskKey := fmt.Sprintf("task:delete:%s", task.ID)
	if err := r.client.HSet(ctx, taskKey, map[string]any{
		"data":       data,
		"status":     task.Status,
		"created_at": task.CreatedAt.Unix(),
	}).Err(); err != nil {
		return fmt.Errorf("store delete task metadata: %w", err)
	}

	r.logger.Info(ctx, "Delete task enqueued successfully",
		observability.String("task_id", task.ID),
		observability.String("object_key", task.ObjectKey))

	return nil
}

// DequeueDeleteTask 出队删除任务（阻塞式）
func (r *RedisQueueRepository) DequeueDeleteTask(ctx context.Context, timeout time.Duration) (*models.DeleteTask, error) {
	// 使用BRPOP阻塞式出队
	result, err := r.client.BRPop(ctx, timeout, "queue:delete_tasks").Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 超时，队列为空
		}
		return nil, fmt.Errorf("dequeue delete task: %w", err)
	}

	if len(result) < 2 {
		return nil, fmt.Errorf("invalid redis response format")
	}

	var task models.DeleteTask
	if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
		return nil, fmt.Errorf("unmarshal delete task: %w", err)
	}

	r.logger.Info(ctx, "Delete task dequeued successfully",
		observability.String("task_id", task.ID),
		observability.String("object_key", task.ObjectKey))

	return &task, nil
}

// UpdateDeleteTaskStatus 更新删除任务状态
func (r *RedisQueueRepository) UpdateDeleteTaskStatus(ctx context.Context, taskID string, status models.TaskStatus, errorMsg string) error {
	taskKey := fmt.Sprintf("task:delete:%s", taskID)

	fields := map[string]any{
		"status":     status,
		"updated_at": time.Now().Unix(),
	}
	if errorMsg != "" {
		fields["error"] = errorMsg
	}

	if err := r.client.HMSet(ctx, taskKey, fields).Err(); err != nil {
		return fmt.Errorf("update delete task status: %w", err)
	}

	// 设置过期时间（7天）
	r.client.Expire(ctx, taskKey, 7*24*time.Hour)

	r.logger.Info(ctx, "Delete task status updated",
		observability.String("task_id", taskID),
		observability.String("status", string(status)))

	return nil
}

// SaveTask队列操作

// EnqueueSaveTask 入队保存任务
func (r *RedisQueueRepository) EnqueueSaveTask(ctx context.Context, task *models.SaveTask) error {
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("marshal save task: %w", err)
	}

	// 使用Redis List作为队列，LPUSH入队
	if err := r.client.LPush(ctx, "queue:save_tasks", data).Err(); err != nil {
		return fmt.Errorf("enqueue save task: %w", err)
	}

	// 同时存储到哈希表中用于状态查询和更新
	taskKey := fmt.Sprintf("task:save:%s", task.ID)
	if err := r.client.HSet(ctx, taskKey, map[string]any{
		"data":       data,
		"status":     task.Status,
		"created_at": task.CreatedAt.Unix(),
	}).Err(); err != nil {
		return fmt.Errorf("store save task metadata: %w", err)
	}

	r.logger.Info(ctx, "Save task enqueued successfully",
		observability.String("task_id", task.ID),
		observability.String("object_key", task.ObjectKey))

	return nil
}

// DequeueSaveTask 出队保存任务（阻塞式）
func (r *RedisQueueRepository) DequeueSaveTask(ctx context.Context, timeout time.Duration) (*models.SaveTask, error) {
	// 使用BRPOP阻塞式出队
	result, err := r.client.BRPop(ctx, timeout, "queue:save_tasks").Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 超时，队列为空
		}
		return nil, fmt.Errorf("dequeue save task: %w", err)
	}

	if len(result) < 2 {
		return nil, fmt.Errorf("invalid redis response format")
	}

	var task models.SaveTask
	if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
		return nil, fmt.Errorf("unmarshal save task: %w", err)
	}

	r.logger.Info(ctx, "Save task dequeued successfully",
		observability.String("task_id", task.ID),
		observability.String("object_key", task.ObjectKey))

	return &task, nil
}

// UpdateSaveTaskStatus 更新保存任务状态
func (r *RedisQueueRepository) UpdateSaveTaskStatus(ctx context.Context, taskID string, status models.TaskStatus, errorMsg string) error {
	taskKey := fmt.Sprintf("task:save:%s", taskID)

	fields := map[string]any{
		"status":     status,
		"updated_at": time.Now().Unix(),
	}
	if errorMsg != "" {
		fields["error"] = errorMsg
	}

	if err := r.client.HMSet(ctx, taskKey, fields).Err(); err != nil {
		return fmt.Errorf("update save task status: %w", err)
	}

	// 设置过期时间（7天）
	r.client.Expire(ctx, taskKey, 7*24*time.Hour)

	r.logger.Info(ctx, "Save task status updated",
		observability.String("task_id", taskID),
		observability.String("status", string(status)))

	return nil
}

// 队列统计操作

// GetDeleteQueueLength 获取删除任务队列长度
func (r *RedisQueueRepository) GetDeleteQueueLength(ctx context.Context) (int64, error) {
	length, err := r.client.LLen(ctx, "queue:delete_tasks").Result()
	if err != nil {
		return 0, fmt.Errorf("get delete queue length: %w", err)
	}
	return length, nil
}

// GetSaveQueueLength 获取保存任务队列长度
func (r *RedisQueueRepository) GetSaveQueueLength(ctx context.Context) (int64, error) {
	length, err := r.client.LLen(ctx, "queue:save_tasks").Result()
	if err != nil {
		return 0, fmt.Errorf("get save queue length: %w", err)
	}
	return length, nil
}

// HealthCheck 健康检查
func (r *RedisQueueRepository) HealthCheck(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close 关闭连接
func (r *RedisQueueRepository) Close() error {
	return r.client.Close()
}
