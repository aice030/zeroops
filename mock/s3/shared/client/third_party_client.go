package client

import (
	"context"
	"fmt"
	"io"
	"mocks3/shared/middleware/consul"
	"mocks3/shared/models"
	"mocks3/shared/observability"
	"time"
)

// ThirdPartyClient 第三方服务客户端
type ThirdPartyClient struct {
	*BaseHTTPClient
}

// NewThirdPartyClient 创建第三方服务客户端
func NewThirdPartyClient(baseURL string, timeout time.Duration, logger *observability.Logger) *ThirdPartyClient {
	return &ThirdPartyClient{
		BaseHTTPClient: NewBaseHTTPClient(baseURL, timeout, "third-party-client", logger),
	}
}

// NewThirdPartyClientWithConsul 创建支持Consul服务发现的第三方服务客户端
func NewThirdPartyClientWithConsul(consulClient consul.ConsulClient, timeout time.Duration, logger *observability.Logger) *ThirdPartyClient {
	ctx := context.Background()
	baseURL := getServiceURL(ctx, consulClient, "third-party-service", "http://localhost:8084", logger)
	return &ThirdPartyClient{
		BaseHTTPClient: NewBaseHTTPClient(baseURL, timeout, "third-party-client", logger),
	}
}

// GetObject 获取对象
func (c *ThirdPartyClient) GetObject(ctx context.Context, bucket, key string) (*models.Object, error) {
	path := fmt.Sprintf("/api/v1/objects/%s/%s", PathEscape(bucket), PathEscape(key))

	// 直接使用DoRequest获取响应以处理二进制数据
	opts := RequestOptions{
		Method: "GET",
		Path:   path,
	}

	resp, err := c.DoRequest(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("third-party service returned status %d", resp.StatusCode)
	}

	// 读取响应数据
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	// 构建对象
	object := &models.Object{
		Key:         key,
		Bucket:      bucket,
		Data:        data,
		Size:        int64(len(data)),
		ContentType: resp.Header.Get("Content-Type"),
		MD5Hash:     resp.Header.Get("ETag"),
		Headers:     make(map[string]string),
		Tags:        make(map[string]string),
	}

	// 复制相关响应头
	for k, v := range resp.Header {
		if len(v) > 0 && (k == "X-Third-Party-Source" || k == "X-Generated-At") {
			object.Headers[k] = v[0]
		}
	}

	// 如果有第三方来源标识，设置标签
	if resp.Header.Get("X-Third-Party-Source") != "" {
		object.Tags["source"] = "third-party"
	}

	return object, nil
}

// HealthCheck 健康检查
func (c *ThirdPartyClient) HealthCheck(ctx context.Context) error {
	return c.BaseHTTPClient.HealthCheck(ctx)
}
