package client

import (
	"context"
	"fmt"
	"mocks3/shared/models"
	"time"
)

// ThirdPartyClient 第三方服务客户端
type ThirdPartyClient struct {
	*BaseHTTPClient
}

// NewThirdPartyClient 创建第三方服务客户端
func NewThirdPartyClient(baseURL string, timeout time.Duration) *ThirdPartyClient {
	return &ThirdPartyClient{
		BaseHTTPClient: NewBaseHTTPClient(baseURL, timeout),
	}
}

// GetObject 获取对象
func (c *ThirdPartyClient) GetObject(ctx context.Context, bucket, key string) (*models.Object, error) {
	path := fmt.Sprintf("/objects/%s/%s", PathEscape(bucket), PathEscape(key))
	var object models.Object
	err := c.Get(ctx, path, nil, &object)
	if err != nil {
		return nil, fmt.Errorf("object not found: %s/%s", bucket, key)
	}
	return &object, nil
}

// HealthCheck 健康检查
func (c *ThirdPartyClient) HealthCheck(ctx context.Context) error {
	return c.BaseHTTPClient.HealthCheck(ctx)
}
