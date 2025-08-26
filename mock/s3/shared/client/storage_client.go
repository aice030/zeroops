package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mocks3/shared/middleware/consul"
	"mocks3/shared/models"
	"mocks3/shared/observability"
	"net/http"
	"time"
)

// StorageClient 存储服务客户端
type StorageClient struct {
	*BaseHTTPClient
}

// NewStorageClient 创建存储服务客户端
func NewStorageClient(baseURL string, timeout time.Duration, logger *observability.Logger) *StorageClient {
	return &StorageClient{
		BaseHTTPClient: NewBaseHTTPClient(baseURL, timeout, "storage-client", logger),
	}
}

// NewStorageClientWithConsul 创建支持Consul服务发现的存储服务客户端
func NewStorageClientWithConsul(consulClient consul.ConsulClient, timeout time.Duration, logger *observability.Logger) *StorageClient {
	ctx := context.Background()
	baseURL := getServiceURL(ctx, consulClient, "storage-service", "http://localhost:8082", logger)
	return &StorageClient{
		BaseHTTPClient: NewBaseHTTPClient(baseURL, timeout, "storage-client", logger),
	}
}

// WriteObject 写入对象（完整流程：保存文件 + 保存元数据）
func (c *StorageClient) WriteObject(ctx context.Context, object *models.Object) error {
	req := &models.UploadRequest{
		Key:         object.Key,
		Bucket:      object.Bucket,
		ContentType: object.ContentType,
		Headers:     object.Headers,
		Tags:        object.Tags,
		Data:        object.Data,
	}

	var uploadResp models.UploadResponse
	if err := c.Post(ctx, "/api/v1/objects", req, &uploadResp); err != nil {
		return err
	}

	if !uploadResp.Success {
		return fmt.Errorf("upload failed: %s", uploadResp.Message)
	}

	object.ID = uploadResp.ObjectID
	object.MD5Hash = uploadResp.MD5Hash
	return nil
}

// WriteObjectToStorage 仅写入到存储节点（内部API，用于队列任务处理）
func (c *StorageClient) WriteObjectToStorage(ctx context.Context, object *models.Object) error {
	req := &models.UploadRequest{
		Key:         object.Key,
		Bucket:      object.Bucket,
		ContentType: object.ContentType,
		Headers:     object.Headers,
		Tags:        object.Tags,
		Data:        object.Data,
	}

	var uploadResp models.UploadResponse
	if err := c.Post(ctx, "/api/v1/internal/objects", req, &uploadResp); err != nil {
		return err
	}

	if !uploadResp.Success {
		return fmt.Errorf("upload to storage failed: %s", uploadResp.Message)
	}

	object.ID = uploadResp.ObjectID
	object.MD5Hash = uploadResp.MD5Hash
	return nil
}

// WriteObjectStream 流式写入对象
func (c *StorageClient) WriteObjectStream(ctx context.Context, bucket, key, contentType string, data io.Reader, size int64) (*models.UploadResponse, error) {
	path := fmt.Sprintf("/api/v1/objects/%s/%s", PathEscape(bucket), PathEscape(key))

	req, err := http.NewRequestWithContext(ctx, "PUT", c.baseURL+path, data)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Content-Length", fmt.Sprintf("%d", size))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var uploadResp models.UploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &uploadResp, nil
}

// ReadObject 读取对象
func (c *StorageClient) ReadObject(ctx context.Context, bucket, key string) (*models.Object, error) {
	path := fmt.Sprintf("/api/v1/objects/%s/%s", PathEscape(bucket), PathEscape(key))

	resp, err := c.DoRequest(ctx, RequestOptions{
		Method: "GET",
		Path:   path,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("object not found: %s/%s", bucket, key)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	object := &models.Object{
		Key:         key,
		Bucket:      bucket,
		Data:        data,
		Size:        int64(len(data)),
		ContentType: resp.Header.Get("Content-Type"),
		MD5Hash:     resp.Header.Get("Content-MD5"),
		Headers:     make(map[string]string),
	}

	// 复制所有响应头
	for k, v := range resp.Header {
		if len(v) > 0 {
			object.Headers[k] = v[0]
		}
	}

	return object, nil
}

// DeleteObject 删除对象（完整流程：删除元数据 + 异步删除文件）
func (c *StorageClient) DeleteObject(ctx context.Context, bucket, key string) error {
	path := fmt.Sprintf("/api/v1/objects/%s/%s", PathEscape(bucket), PathEscape(key))
	return c.Delete(ctx, path)
}

// DeleteObjectFromStorage 仅从存储节点删除文件（内部API，用于队列任务处理）
func (c *StorageClient) DeleteObjectFromStorage(ctx context.Context, bucket, key string) error {
	path := fmt.Sprintf("/api/v1/internal/objects/%s/%s", PathEscape(bucket), PathEscape(key))
	return c.Delete(ctx, path)
}

// ListObjects 列出对象
func (c *StorageClient) ListObjects(ctx context.Context, req *models.ListObjectsRequest) (*models.ListObjectsResponse, error) {
	queryParams := BuildQueryParams(map[string]any{
		"bucket":      req.Bucket,
		"prefix":      req.Prefix,
		"delimiter":   req.Delimiter,
		"max_keys":    req.MaxKeys,
		"start_after": req.StartAfter,
	})

	var listResp models.ListObjectsResponse
	err := c.Get(ctx, "/api/v1/objects", queryParams, &listResp)
	return &listResp, err
}

// HealthCheck 健康检查
func (c *StorageClient) HealthCheck(ctx context.Context) error {
	return c.BaseHTTPClient.HealthCheck(ctx)
}
