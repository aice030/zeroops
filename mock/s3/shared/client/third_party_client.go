package client

import (
	"context"
	"fmt"
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

// GetObject 获取对象
func (c *ThirdPartyClient) GetObject(ctx context.Context, bucket, key string) (*models.Object, error) {
	path := fmt.Sprintf("/api/v1/objects/%s/%s", PathEscape(bucket), PathEscape(key))
	var object models.Object
	err := c.Get(ctx, path, nil, &object)
	if err != nil {
		return nil, err
	}
	return &object, nil
}

// HealthCheck 健康检查
func (c *ThirdPartyClient) HealthCheck(ctx context.Context) error {
	return c.BaseHTTPClient.HealthCheck(ctx)
}
