package client

import (
	"context"
	"fmt"
	"mocks3/shared/models"
	"net/http"
	"time"
)

// QueueClient 队列服务客户端
type QueueClient struct {
	*BaseHTTPClient
}

// NewQueueClient 创建队列服务客户端
func NewQueueClient(baseURL string, timeout time.Duration) *QueueClient {
	return &QueueClient{
		BaseHTTPClient: NewBaseHTTPClient(baseURL, timeout),
	}
}

// EnqueueDeleteTask 入队删除任务
func (c *QueueClient) EnqueueDeleteTask(ctx context.Context, task *models.DeleteTask) error {
	return c.PostExpectStatus(ctx, "/delete-tasks", task, http.StatusCreated)
}

// DequeueDeleteTask 出队删除任务
func (c *QueueClient) DequeueDeleteTask(ctx context.Context) (*models.DeleteTask, error) {
	resp, err := c.DoRequest(ctx, RequestOptions{
		Method: "GET",
		Path:   "/delete-tasks/dequeue",
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil, nil // 队列为空
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var task models.DeleteTask
	if err := c.DoRequestWithJSON(ctx, RequestOptions{
		Method: "GET",
		Path:   "/delete-tasks/dequeue",
	}, &task); err != nil {
		return nil, err
	}

	return &task, nil
}

// GetQueueLength 获取队列长度
func (c *QueueClient) GetQueueLength(ctx context.Context) (int64, error) {
	var result struct {
		Length int64 `json:"length"`
	}
	err := c.Get(ctx, "/delete-tasks/length", nil, &result)
	return result.Length, err
}

// UpdateDeleteTaskStatus 更新删除任务状态
func (c *QueueClient) UpdateDeleteTaskStatus(ctx context.Context, taskID string, status models.TaskStatus, errorMsg string) error {
	path := fmt.Sprintf("/delete-tasks/%s/status", PathEscape(taskID))
	req := map[string]any{
		"status": status,
	}
	if errorMsg != "" {
		req["error"] = errorMsg
	}
	return c.PutExpectStatus(ctx, path, req, http.StatusOK)
}

// HealthCheck 健康检查
func (c *QueueClient) HealthCheck(ctx context.Context) error {
	return c.BaseHTTPClient.HealthCheck(ctx)
}
