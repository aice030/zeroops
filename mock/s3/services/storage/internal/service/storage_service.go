package service

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"mocks3/services/storage/internal/repository"
	"mocks3/shared/client"
	"mocks3/shared/models"
	"mocks3/shared/observability"
	"time"
)

// StorageService Storage服务实现
type StorageService struct {
	repo             *repository.FileStorageRepository
	metadataClient   *client.MetadataClient
	queueClient      *client.QueueClient
	thirdPartyClient *client.ThirdPartyClient
	logger           *observability.Logger
}

// NewStorageService 创建Storage服务
func NewStorageService(
	repo *repository.FileStorageRepository,
	metadataClient *client.MetadataClient,
	queueClient *client.QueueClient,
	thirdPartyClient *client.ThirdPartyClient,
	logger *observability.Logger,
) *StorageService {
	return &StorageService{
		repo:             repo,
		metadataClient:   metadataClient,
		queueClient:      queueClient,
		thirdPartyClient: thirdPartyClient,
		logger:           logger,
	}
}

// WriteObject 写入对象 - 存入存储节点，保存元数据
func (s *StorageService) WriteObject(ctx context.Context, object *models.Object) error {
	s.logger.Info(ctx, "Writing object to storage",
		observability.String("bucket", object.Bucket),
		observability.String("key", object.Key),
		observability.Int64("size", object.Size))

	// 计算MD5哈希
	if object.MD5Hash == "" && object.Data != nil {
		hash := md5.Sum(object.Data)
		object.MD5Hash = fmt.Sprintf("%x", hash)
	}

	// 1. 写入到存储节点
	dataReader := bytes.NewReader(object.Data)
	if err := s.repo.WriteObject(ctx, object.Bucket, object.Key, dataReader, object.Size); err != nil {
		s.logger.Error(ctx, "Failed to write object to storage nodes",
			observability.String("bucket", object.Bucket),
			observability.String("key", object.Key),
			observability.Error(err))
		return fmt.Errorf("write to storage nodes: %w", err)
	}

	// 2. 保存元数据到Metadata Service
	metadata := &models.Metadata{
		Bucket:      object.Bucket,
		Key:         object.Key,
		Size:        object.Size,
		ContentType: object.ContentType,
		MD5Hash:     object.MD5Hash,
		Status:      "active",
		CreatedAt:   time.Now(),
	}

	if err := s.metadataClient.SaveMetadata(ctx, metadata); err != nil {
		s.logger.Error(ctx, "Failed to save metadata",
			observability.String("bucket", object.Bucket),
			observability.String("key", object.Key),
			observability.Error(err))

		// 如果元数据保存失败，异步删除已写入的文件
		s.scheduleCleanup(ctx, object.Bucket, object.Key)
		return fmt.Errorf("save metadata: %w", err)
	}

	s.logger.Info(ctx, "Object written successfully",
		observability.String("bucket", object.Bucket),
		observability.String("key", object.Key),
		observability.String("md5", object.MD5Hash))

	return nil
}

// ReadObject 读取对象 - 读取元数据，从存储节点获取，失败则从第三方获取并保存
func (s *StorageService) ReadObject(ctx context.Context, bucket, key string) (*models.Object, error) {
	s.logger.Info(ctx, "Reading object from storage",
		observability.String("bucket", bucket),
		observability.String("key", key))

	// 1. 首先读取元数据
	metadata, err := s.metadataClient.GetMetadata(ctx, bucket, key)
	if err != nil {
		s.logger.Error(ctx, "Failed to get metadata",
			observability.String("bucket", bucket),
			observability.String("key", key),
			observability.Error(err))
		return nil, fmt.Errorf("metadata not found: %s/%s", bucket, key)
	}

	// 2. 尝试从存储节点读取文件
	reader, size, err := s.repo.ReadObject(ctx, bucket, key)
	if err == nil {
		defer reader.Close()

		// 读取所有数据
		data, readErr := io.ReadAll(reader)
		if readErr == nil {
			// 成功从本地存储读取
			object := s.buildObjectFromMetadata(metadata, data, size)
			s.logger.Info(ctx, "Object read from local storage",
				observability.String("bucket", bucket),
				observability.String("key", key),
				observability.String("source", "local"))
			return object, nil
		}
		s.logger.Warn(ctx, "Failed to read data from local storage", observability.Error(readErr))
	}

	s.logger.Warn(ctx, "Local storage failed, trying third-party service",
		observability.String("bucket", bucket),
		observability.String("key", key),
		observability.Error(err))

	// 3. 如果本地存储失败，尝试第三方服务
	if s.thirdPartyClient != nil {
		thirdPartyObject, thirdPartyErr := s.thirdPartyClient.GetObject(ctx, bucket, key)
		if thirdPartyErr == nil {
			s.logger.Info(ctx, "Object retrieved from third-party service, scheduling save to local storage",
				observability.String("bucket", bucket),
				observability.String("key", key),
				observability.String("source", "third-party"))

			// 通过Queue Service异步保存到本地存储和更新元数据
			s.scheduleThirdPartySaveTask(ctx, thirdPartyObject)

			return thirdPartyObject, nil
		}
		s.logger.Warn(ctx, "Third-party service also failed",
			observability.Error(thirdPartyErr))
	}

	// 4. 所有方式都失败
	s.logger.Error(ctx, "Failed to read object from all sources",
		observability.String("bucket", bucket),
		observability.String("key", key),
		observability.Error(err))

	return nil, fmt.Errorf("object not found in storage or third-party service: %s/%s", bucket, key)
}

// buildObjectFromMetadata 从元数据构建对象
func (s *StorageService) buildObjectFromMetadata(metadata *models.Metadata, data []byte, size int64) *models.Object {
	return &models.Object{
		ID:          metadata.GetID(),
		Bucket:      metadata.Bucket,
		Key:         metadata.Key,
		Data:        data,
		Size:        size,
		ContentType: metadata.ContentType,
		MD5Hash:     metadata.MD5Hash,
		Headers:     make(map[string]string),
		Tags:        make(map[string]string),
	}
}

// scheduleThirdPartySaveTask 调度第三方对象保存任务到Queue Service
func (s *StorageService) scheduleThirdPartySaveTask(ctx context.Context, object *models.Object) {
	if s.queueClient == nil {
		s.logger.Warn(ctx, "Queue client not available, skipping third-party save task",
			observability.String("bucket", object.Bucket),
			observability.String("key", object.Key))
		return
	}

	saveTask := &models.SaveTask{
		ObjectKey: object.Bucket + "/" + object.Key,
		Object:    object,
		CreatedAt: time.Now(),
		Status:    models.TaskStatusPending,
	}
	saveTask.GenerateID()

	if err := s.queueClient.EnqueueSaveTask(ctx, saveTask); err != nil {
		s.logger.Error(ctx, "Failed to schedule third-party save task",
			observability.String("bucket", object.Bucket),
			observability.String("key", object.Key),
			observability.Error(err))
	} else {
		s.logger.Info(ctx, "Third-party save task scheduled successfully",
			observability.String("bucket", object.Bucket),
			observability.String("key", object.Key),
			observability.String("task_id", saveTask.ID))
	}
}

// DeleteObject 删除对象 - 删除元数据，异步删除存储节点文件
func (s *StorageService) DeleteObject(ctx context.Context, bucket, key string) error {
	s.logger.Info(ctx, "Deleting object from storage",
		observability.String("bucket", bucket),
		observability.String("key", key))

	// 1. 首先删除元数据
	if err := s.metadataClient.DeleteMetadata(ctx, bucket, key); err != nil {
		s.logger.Error(ctx, "Failed to delete metadata",
			observability.String("bucket", bucket),
			observability.String("key", key),
			observability.Error(err))
		return fmt.Errorf("delete metadata: %w", err)
	}

	s.logger.Info(ctx, "Metadata deleted successfully",
		observability.String("bucket", bucket),
		observability.String("key", key))

	// 2. 通过Queue Service异步删除存储节点中的文件
	s.scheduleDeleteTask(ctx, bucket, key)

	return nil
}

// scheduleDeleteTask 调度删除存储节点任务到Queue Service
func (s *StorageService) scheduleDeleteTask(ctx context.Context, bucket, key string) {
	if s.queueClient == nil {
		s.logger.Warn(ctx, "Queue client not available, skipping delete task",
			observability.String("bucket", bucket),
			observability.String("key", key))
		return
	}

	deleteTask := &models.DeleteTask{
		ObjectKey: bucket + "/" + key,
		CreatedAt: time.Now(),
		Status:    models.TaskStatusPending,
	}
	deleteTask.GenerateID()

	if err := s.queueClient.EnqueueDeleteTask(ctx, deleteTask); err != nil {
		s.logger.Error(ctx, "Failed to schedule delete task",
			observability.String("bucket", bucket),
			observability.String("key", key),
			observability.Error(err))
	} else {
		s.logger.Info(ctx, "Delete task scheduled successfully",
			observability.String("bucket", bucket),
			observability.String("key", key),
			observability.String("task_id", deleteTask.ID))
	}
}

// ListObjects 列出对象
func (s *StorageService) ListObjects(ctx context.Context, req *models.ListObjectsRequest) (*models.ListObjectsResponse, error) {
	s.logger.Info(ctx, "Listing objects",
		observability.String("bucket", req.Bucket),
		observability.String("prefix", req.Prefix))

	// 通过Metadata Service获取对象列表
	metadataList, err := s.metadataClient.ListMetadata(ctx, req.Bucket, req.Prefix, req.MaxKeys, 0)
	if err != nil {
		s.logger.Error(ctx, "Failed to list metadata",
			observability.String("bucket", req.Bucket),
			observability.Error(err))
		return nil, fmt.Errorf("list metadata: %w", err)
	}

	// 转换为对象信息
	objects := make([]models.ObjectInfo, 0, len(metadataList))
	for _, metadata := range metadataList {
		objects = append(objects, models.ObjectInfo{
			ID:          metadata.GetID(),
			Key:         metadata.Key,
			Bucket:      metadata.Bucket,
			Size:        metadata.Size,
			ContentType: metadata.ContentType,
			MD5Hash:     metadata.MD5Hash,
			CreatedAt:   metadata.CreatedAt,
		})
	}

	response := &models.ListObjectsResponse{
		Bucket:      req.Bucket,
		Prefix:      req.Prefix,
		Objects:     objects,
		IsTruncated: len(objects) >= req.MaxKeys,
		MaxKeys:     req.MaxKeys,
		Count:       len(objects),
	}

	s.logger.Info(ctx, "Objects listed successfully",
		observability.String("bucket", req.Bucket),
		observability.Int("count", len(objects)))

	return response, nil
}

// GetStats 获取统计信息
func (s *StorageService) GetStats(ctx context.Context) (map[string]any, error) {
	s.logger.Info(ctx, "Getting storage stats")

	// 获取节点状态
	nodeStats, err := s.repo.GetNodeStats(ctx)
	if err != nil {
		s.logger.Error(ctx, "Failed to get node stats", observability.Error(err))
		return nil, fmt.Errorf("get node stats: %w", err)
	}

	// 计算总体统计
	var totalUsedSpace int64
	healthyNodes := 0
	for _, stat := range nodeStats {
		totalUsedSpace += stat.UsedSpace
		if stat.Status == "healthy" {
			healthyNodes++
		}
	}

	stats := map[string]any{
		"service":          "storage-service",
		"nodes":            nodeStats,
		"total_nodes":      len(nodeStats),
		"healthy_nodes":    healthyNodes,
		"total_used_space": totalUsedSpace,
		"timestamp":        time.Now(),
	}

	s.logger.Info(ctx, "Storage stats retrieved",
		observability.Int("total_nodes", len(nodeStats)),
		observability.Int("healthy_nodes", healthyNodes),
		observability.Int64("total_used_space", totalUsedSpace))

	return stats, nil
}

// HealthCheck 健康检查
func (s *StorageService) HealthCheck(ctx context.Context) error {
	// 检查节点状态
	nodeStats, err := s.repo.GetNodeStats(ctx)
	if err != nil {
		return fmt.Errorf("get node stats: %w", err)
	}

	// 至少要有一个健康节点
	healthyNodes := 0
	for _, stat := range nodeStats {
		if stat.Status == "healthy" {
			healthyNodes++
		}
	}

	if healthyNodes == 0 {
		return fmt.Errorf("no healthy storage nodes available")
	}

	return nil
}

// scheduleCleanup 通过Queue Service调度清理任务（用于删除失败写入的文件）
func (s *StorageService) scheduleCleanup(ctx context.Context, bucket, key string) {
	if s.queueClient == nil {
		s.logger.Warn(ctx, "Queue client not available, cannot schedule cleanup task",
			observability.String("bucket", bucket),
			observability.String("key", key))
		return
	}

	deleteTask := &models.DeleteTask{
		ObjectKey: bucket + "/" + key,
		CreatedAt: time.Now(),
		Status:    models.TaskStatusPending,
	}
	deleteTask.GenerateID()

	if err := s.queueClient.EnqueueDeleteTask(ctx, deleteTask); err != nil {
		s.logger.Error(ctx, "Failed to schedule cleanup task",
			observability.String("bucket", bucket),
			observability.String("key", key),
			observability.Error(err))
	} else {
		s.logger.Info(ctx, "Cleanup task scheduled successfully",
			observability.String("bucket", bucket),
			observability.String("key", key),
			observability.String("task_id", deleteTask.ID))
	}
}
