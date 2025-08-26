package middleware

import (
	"context"
	"mocks3/shared/models"

	"github.com/gin-gonic/gin"
)

// ConsulClient Consul客户端接口
type ConsulClient interface {
	// 服务注册
	RegisterService(ctx context.Context, service *models.ServiceInfo) error
	DeregisterService(ctx context.Context, serviceID string) error

	// 服务发现
	GetHealthyServices(ctx context.Context, serviceName string) ([]*models.ServiceInfo, error)

	// 配置管理
	GetConfig(ctx context.Context, key string) (string, error)
	SetConfig(ctx context.Context, key, value string) error
}

// HTTP中间件（暂时空实现）
func ServiceDiscoveryMiddleware(consulClient ConsulClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

func HealthCheckMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

// 工具函数（将来实现）
func CreateConsulClient(address string) (ConsulClient, error) {
	panic("not implemented")
}

func RegisterService(ctx context.Context, consulClient ConsulClient,
	serviceName, address string, port int) error {
	service := &models.ServiceInfo{
		Name:    serviceName,
		Address: address,
		Port:    port,
		Health:  models.HealthStatusHealthy,
	}
	return consulClient.RegisterService(ctx, service)
}
