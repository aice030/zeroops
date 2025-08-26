package worker

import (
	"context"
	"fmt"
	"mocks3/shared/client"
	"mocks3/shared/interfaces"
	"mocks3/shared/models"
	"mocks3/shared/observability"
	"strings"
)

// StorageDeleteProcessor 删除任务处理器 - 通过HTTP调用Storage Service
type StorageDeleteProcessor struct {
	storageClient *client.StorageClient
	logger        *observability.Logger
}

// NewStorageDeleteProcessor 创建删除任务处理器
func NewStorageDeleteProcessor(storageClient *client.StorageClient, logger *observability.Logger) *StorageDeleteProcessor {
	return &StorageDeleteProcessor{
		storageClient: storageClient,
		logger:        logger,
	}
}

// ProcessDeleteTask 处理删除任务
func (p *StorageDeleteProcessor) ProcessDeleteTask(ctx context.Context, task *models.DeleteTask) error {
	p.logger.Info(ctx, "Processing delete task via Storage Service",
		observability.String("task_id", task.ID),
		observability.String("object_key", task.ObjectKey))

	// 解析对象键获取bucket和key
	bucket, key, err := parseObjectKey(task.ObjectKey)
	if err != nil {
		return fmt.Errorf("parse object key: %w", err)
	}

	// 通过StorageClient调用Storage Service的内部删除API
	// 这个API只删除存储节点中的文件，不会触发队列任务
	err = p.storageClient.DeleteObjectFromStorage(ctx, bucket, key)
	if err != nil {
		p.logger.Error(ctx, "Failed to delete object from storage nodes",
			observability.String("bucket", bucket),
			observability.String("key", key),
			observability.Error(err))
		return fmt.Errorf("delete object from storage: %w", err)
	}

	p.logger.Info(ctx, "Object deleted via Storage Service successfully",
		observability.String("bucket", bucket),
		observability.String("key", key))

	return nil
}

// parseObjectKey 解析对象键，格式为 "bucket/key"
func parseObjectKey(objectKey string) (bucket, key string, err error) {
	parts := strings.SplitN(objectKey, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid object key format: %s", objectKey)
	}
	return parts[0], parts[1], nil
}

// 确保处理器实现了接口
var _ interfaces.DeleteTaskProcessor = (*StorageDeleteProcessor)(nil)
