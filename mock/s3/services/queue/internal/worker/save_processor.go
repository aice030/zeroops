package worker

import (
	"context"
	"fmt"
	"mocks3/shared/client"
	"mocks3/shared/interfaces"
	"mocks3/shared/models"
	"mocks3/shared/observability"
	"time"
)

// StorageSaveProcessor 保存任务处理器 - 通过HTTP调用Storage Service和Metadata Service
type StorageSaveProcessor struct {
	storageClient  *client.StorageClient
	metadataClient *client.MetadataClient
	logger         *observability.Logger
}

// NewStorageSaveProcessor 创建保存任务处理器
func NewStorageSaveProcessor(
	storageClient *client.StorageClient,
	metadataClient *client.MetadataClient,
	logger *observability.Logger,
) *StorageSaveProcessor {
	return &StorageSaveProcessor{
		storageClient:  storageClient,
		metadataClient: metadataClient,
		logger:         logger,
	}
}

// ProcessSaveTask 处理保存任务
func (p *StorageSaveProcessor) ProcessSaveTask(ctx context.Context, task *models.SaveTask) error {
	p.logger.Info(ctx, "Processing save task via Storage and Metadata Services",
		observability.String("task_id", task.ID),
		observability.String("object_key", task.ObjectKey))

	if task.Object == nil {
		return fmt.Errorf("task object is nil")
	}

	// 1. 通过StorageClient保存到存储节点
	// 使用内部API，只保存文件到存储节点，不触发元数据保存
	err := p.storageClient.WriteObjectToStorage(ctx, task.Object)
	if err != nil {
		p.logger.Error(ctx, "Failed to save object to storage nodes",
			observability.String("bucket", task.Object.Bucket),
			observability.String("key", task.Object.Key),
			observability.Error(err))
		return fmt.Errorf("save object to storage: %w", err)
	}

	// 2. 通过MetadataClient保存/更新元数据
	metadata := &models.Metadata{
		Bucket:      task.Object.Bucket,
		Key:         task.Object.Key,
		Size:        task.Object.Size,
		ContentType: task.Object.ContentType,
		MD5Hash:     task.Object.MD5Hash,
		Status:      "active",
		CreatedAt:   time.Now(),
	}

	err = p.metadataClient.SaveMetadata(ctx, metadata)
	if err != nil {
		p.logger.Error(ctx, "Failed to save metadata for third-party object",
			observability.String("bucket", task.Object.Bucket),
			observability.String("key", task.Object.Key),
			observability.Error(err))

		// 如果元数据保存失败，考虑回滚存储节点中的文件
		// 但这里我们不做回滚，因为这可能是临时的网络问题
		// 可以在后续的数据一致性检查中处理这种情况
		return fmt.Errorf("save metadata: %w", err)
	}

	p.logger.Info(ctx, "Third-party object saved to local storage successfully",
		observability.String("bucket", task.Object.Bucket),
		observability.String("key", task.Object.Key))

	return nil
}

// 确保处理器实现了接口
var _ interfaces.SaveTaskProcessor = (*StorageSaveProcessor)(nil)
