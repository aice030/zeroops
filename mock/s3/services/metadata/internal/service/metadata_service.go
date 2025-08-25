package service

import (
	"context"
	"fmt"
	"mocks3/services/metadata/internal/repository"
	"mocks3/shared/models"
	"mocks3/shared/observability"
	"time"
)

// MetadataService 元数据服务实现
type MetadataService struct {
	repo   *repository.PostgreSQLRepository
	logger *observability.Logger
}

// NewMetadataService 创建元数据服务
func NewMetadataService(repo *repository.PostgreSQLRepository, logger *observability.Logger) *MetadataService {
	return &MetadataService{
		repo:   repo,
		logger: logger,
	}
}

// SaveMetadata 保存元数据
func (s *MetadataService) SaveMetadata(ctx context.Context, metadata *models.Metadata) error {
	s.logger.Info(ctx, "Saving metadata",
		observability.String("bucket", metadata.Bucket),
		observability.String("key", metadata.Key),
		observability.Int64("size", metadata.Size))

	// 设置创建时间
	if metadata.CreatedAt.IsZero() {
		metadata.CreatedAt = time.Now()
	}

	// 设置状态
	if metadata.Status == "" {
		metadata.Status = models.StatusActive
	}

	err := s.repo.Create(ctx, metadata)
	if err != nil {
		s.logger.ErrorWithErr(ctx, err, "Failed to save metadata",
			observability.String("bucket", metadata.Bucket),
			observability.String("key", metadata.Key))
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	s.logger.Info(ctx, "Metadata saved successfully",
		observability.String("bucket", metadata.Bucket),
		observability.String("key", metadata.Key))

	return nil
}

// GetMetadata 获取元数据
func (s *MetadataService) GetMetadata(ctx context.Context, bucket, key string) (*models.Metadata, error) {
	s.logger.Debug(ctx, "Getting metadata",
		observability.String("bucket", bucket),
		observability.String("key", key))

	metadata, err := s.repo.GetByKey(ctx, bucket, key)
	if err != nil {
		s.logger.Warn(ctx, "Metadata not found",
			observability.String("bucket", bucket),
			observability.String("key", key),
			observability.Error(err))
		return nil, fmt.Errorf("metadata not found: %s/%s", bucket, key)
	}

	s.logger.Debug(ctx, "Metadata retrieved successfully",
		observability.String("bucket", bucket),
		observability.String("key", key))

	return metadata, nil
}

// UpdateMetadata 更新元数据
func (s *MetadataService) UpdateMetadata(ctx context.Context, metadata *models.Metadata) error {
	s.logger.Info(ctx, "Updating metadata",
		observability.String("bucket", metadata.Bucket),
		observability.String("key", metadata.Key))

	err := s.repo.Update(ctx, metadata)
	if err != nil {
		s.logger.ErrorWithErr(ctx, err, "Failed to update metadata",
			observability.String("bucket", metadata.Bucket),
			observability.String("key", metadata.Key))
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	s.logger.Info(ctx, "Metadata updated successfully",
		observability.String("bucket", metadata.Bucket),
		observability.String("key", metadata.Key))

	return nil
}

// DeleteMetadata 删除元数据
func (s *MetadataService) DeleteMetadata(ctx context.Context, bucket, key string) error {
	s.logger.Info(ctx, "Deleting metadata",
		observability.String("bucket", bucket),
		observability.String("key", key))

	err := s.repo.Delete(ctx, bucket, key)
	if err != nil {
		s.logger.ErrorWithErr(ctx, err, "Failed to delete metadata",
			observability.String("bucket", bucket),
			observability.String("key", key))
		return fmt.Errorf("failed to delete metadata: %w", err)
	}

	s.logger.Info(ctx, "Metadata deleted successfully",
		observability.String("bucket", bucket),
		observability.String("key", key))

	return nil
}

// ListMetadata 列出元数据
func (s *MetadataService) ListMetadata(ctx context.Context, bucket, prefix string, limit, offset int) ([]*models.Metadata, error) {
	s.logger.Debug(ctx, "Listing metadata",
		observability.String("bucket", bucket),
		observability.String("prefix", prefix),
		observability.Int("limit", limit),
		observability.Int("offset", offset))

	metadataList, err := s.repo.List(ctx, bucket, prefix, limit, offset)
	if err != nil {
		s.logger.ErrorWithErr(ctx, err, "Failed to list metadata",
			observability.String("bucket", bucket),
			observability.String("prefix", prefix))
		return nil, fmt.Errorf("failed to list metadata: %w", err)
	}

	s.logger.Debug(ctx, "Metadata listed successfully",
		observability.String("bucket", bucket),
		observability.String("prefix", prefix),
		observability.Int("count", len(metadataList)))

	return metadataList, nil
}

// SearchMetadata 搜索元数据
func (s *MetadataService) SearchMetadata(ctx context.Context, query string, limit int) ([]*models.Metadata, error) {
	s.logger.Debug(ctx, "Searching metadata",
		observability.String("query", query),
		observability.Int("limit", limit))

	metadataList, err := s.repo.Search(ctx, query, limit)
	if err != nil {
		s.logger.ErrorWithErr(ctx, err, "Failed to search metadata",
			observability.String("query", query))
		return nil, fmt.Errorf("failed to search metadata: %w", err)
	}

	s.logger.Debug(ctx, "Metadata search completed",
		observability.String("query", query),
		observability.Int("count", len(metadataList)))

	return metadataList, nil
}

// GetStats 获取统计信息
func (s *MetadataService) GetStats(ctx context.Context) (*models.Stats, error) {
	s.logger.Debug(ctx, "Getting stats")

	stats, err := s.repo.GetStats(ctx)
	if err != nil {
		s.logger.ErrorWithErr(ctx, err, "Failed to get stats")
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	s.logger.Debug(ctx, "Stats retrieved successfully",
		observability.Int64("total_objects", stats.TotalObjects),
		observability.Int64("total_size", stats.TotalSize))

	return stats, nil
}

// HealthCheck 健康检查
func (s *MetadataService) HealthCheck(ctx context.Context) error {
	s.logger.Debug(ctx, "Performing health check")

	// 简单查询来检测数据库连接
	_, err := s.repo.GetStats(ctx)
	if err != nil {
		s.logger.ErrorWithErr(ctx, err, "Health check failed")
		return fmt.Errorf("health check failed: %w", err)
	}

	s.logger.Debug(ctx, "Health check passed")
	return nil
}
