package client

import (
	"context"
	"fmt"
	"mocks3/shared/middleware/consul"
	"mocks3/shared/models"
	"mocks3/shared/observability"
	"net/http"
	"time"
)

// MetadataClient 元数据服务客户端
type MetadataClient struct {
	*BaseHTTPClient
}

// NewMetadataClient 创建元数据服务客户端
func NewMetadataClient(baseURL string, timeout time.Duration, logger *observability.Logger) *MetadataClient {
	return &MetadataClient{
		BaseHTTPClient: NewBaseHTTPClient(baseURL, timeout, "metadata-client", logger),
	}
}

// NewMetadataClientWithConsul 创建支持Consul服务发现的元数据服务客户端
func NewMetadataClientWithConsul(consulClient consul.ConsulClient, timeout time.Duration, logger *observability.Logger) *MetadataClient {
	ctx := context.Background()
	baseURL := getServiceURL(ctx, consulClient, "metadata-service", "http://localhost:8081", logger)
	return &MetadataClient{
		BaseHTTPClient: NewBaseHTTPClient(baseURL, timeout, "metadata-client", logger),
	}
}

// SaveMetadata 保存元数据
func (c *MetadataClient) SaveMetadata(ctx context.Context, metadata *models.Metadata) error {
	return c.PostExpectStatus(ctx, "/api/v1/metadata", metadata, http.StatusCreated)
}

// GetMetadata 获取元数据
func (c *MetadataClient) GetMetadata(ctx context.Context, bucket, key string) (*models.Metadata, error) {
	path := fmt.Sprintf("/api/v1/metadata/%s/%s", PathEscape(bucket), PathEscape(key))
	var metadata models.Metadata
	err := c.Get(ctx, path, nil, &metadata)
	if err != nil {
		return nil, err
	}
	return &metadata, nil
}

// UpdateMetadata 更新元数据
func (c *MetadataClient) UpdateMetadata(ctx context.Context, metadata *models.Metadata) error {
	path := fmt.Sprintf("/api/v1/metadata/%s/%s", PathEscape(metadata.Bucket), PathEscape(metadata.Key))
	return c.PutExpectStatus(ctx, path, metadata, http.StatusOK)
}

// DeleteMetadata 删除元数据
func (c *MetadataClient) DeleteMetadata(ctx context.Context, bucket, key string) error {
	path := fmt.Sprintf("/api/v1/metadata/%s/%s", PathEscape(bucket), PathEscape(key))
	return c.Delete(ctx, path)
}

// ListMetadata 列出元数据
func (c *MetadataClient) ListMetadata(ctx context.Context, bucket, prefix string, limit, offset int) ([]*models.Metadata, error) {
	queryParams := BuildQueryParams(map[string]any{
		"bucket": bucket,
		"prefix": prefix,
		"limit":  limit,
		"offset": offset,
	})

	var response struct {
		Metadata []*models.Metadata `json:"metadata"`
		Count    int                `json:"count"`
		Bucket   string             `json:"bucket"`
		Prefix   string             `json:"prefix"`
		Limit    int                `json:"limit"`
		Offset   int                `json:"offset"`
	}

	err := c.Get(ctx, "/api/v1/metadata", queryParams, &response)
	return response.Metadata, err
}

// SearchMetadata 搜索元数据
func (c *MetadataClient) SearchMetadata(ctx context.Context, query, bucket string, limit int) ([]*models.Metadata, error) {
	queryParams := BuildQueryParams(map[string]any{
		"q":      query,
		"bucket": bucket,
		"limit":  limit,
	})

	var response struct {
		Query    string             `json:"query"`
		Metadata []*models.Metadata `json:"metadata"`
		Count    int                `json:"count"`
		Limit    int                `json:"limit"`
	}

	err := c.Get(ctx, "/api/v1/metadata/search", queryParams, &response)
	if err != nil {
		return nil, err
	}

	return response.Metadata, nil
}

// GetStats 获取统计信息
func (c *MetadataClient) GetStats(ctx context.Context) (*models.Stats, error) {
	var stats models.Stats
	err := c.Get(ctx, "/api/v1/stats", nil, &stats)
	return &stats, err
}

// HealthCheck 健康检查
func (c *MetadataClient) HealthCheck(ctx context.Context) error {
	return c.BaseHTTPClient.HealthCheck(ctx)
}
