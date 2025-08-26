package consul

import (
	"context"
	"fmt"
	"mocks3/shared/models"
	"mocks3/shared/observability"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
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

// DefaultConsulClient Consul客户端实现
type DefaultConsulClient struct {
	client *api.Client
	logger *observability.Logger
}

// CreateConsulClient 创建Consul客户端
func CreateConsulClient(address string, logger *observability.Logger) (ConsulClient, error) {
	config := api.DefaultConfig()
	config.Address = address

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	return &DefaultConsulClient{
		client: client,
		logger: logger,
	}, nil
}

// RegisterService 注册服务到Consul
func (c *DefaultConsulClient) RegisterService(ctx context.Context, service *models.ServiceInfo) error {
	registration := &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("%s-%s-%d", service.Name, service.Address, service.Port),
		Name:    service.Name,
		Address: service.Address,
		Port:    service.Port,
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health", service.Address, service.Port),
			Interval:                       "30s",
			Timeout:                        "10s",
			DeregisterCriticalServiceAfter: "5m",
		},
		Tags: []string{"mock-s3", "microservice"},
	}

	err := c.client.Agent().ServiceRegister(registration)
	if err != nil {
		if c.logger != nil {
			c.logger.Error(ctx, "Failed to register service",
				observability.String("service_name", service.Name),
				observability.String("address", service.Address),
				observability.Int("port", service.Port),
				observability.Error(err))
		}
		return fmt.Errorf("failed to register service: %w", err)
	}

	if c.logger != nil {
		c.logger.Info(ctx, "Service registered successfully",
			observability.String("service_name", service.Name),
			observability.String("service_id", registration.ID))
	}

	return nil
}

// DeregisterService 注销服务
func (c *DefaultConsulClient) DeregisterService(ctx context.Context, serviceID string) error {
	err := c.client.Agent().ServiceDeregister(serviceID)
	if err != nil {
		if c.logger != nil {
			c.logger.Error(ctx, "Failed to deregister service",
				observability.String("service_id", serviceID),
				observability.Error(err))
		}
		return fmt.Errorf("failed to deregister service: %w", err)
	}

	if c.logger != nil {
		c.logger.Info(ctx, "Service deregistered successfully",
			observability.String("service_id", serviceID))
	}

	return nil
}

// GetHealthyServices 获取健康的服务实例
func (c *DefaultConsulClient) GetHealthyServices(ctx context.Context, serviceName string) ([]*models.ServiceInfo, error) {
	services, _, err := c.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		if c.logger != nil {
			c.logger.Error(ctx, "Failed to get healthy services",
				observability.String("service_name", serviceName),
				observability.Error(err))
		}
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	var result []*models.ServiceInfo
	for _, service := range services {
		info := &models.ServiceInfo{
			ID:      service.Service.ID,
			Name:    service.Service.Service,
			Address: service.Service.Address,
			Port:    service.Service.Port,
			Health:  models.HealthStatusHealthy,
			Tags:    service.Service.Tags,
		}
		result = append(result, info)
	}

	if c.logger != nil {
		c.logger.Debug(ctx, "Retrieved healthy services",
			observability.String("service_name", serviceName),
			observability.Int("count", len(result)))
	}

	return result, nil
}

// GetConfig 从Consul KV获取配置
func (c *DefaultConsulClient) GetConfig(ctx context.Context, key string) (string, error) {
	pair, _, err := c.client.KV().Get(key, nil)
	if err != nil {
		if c.logger != nil {
			c.logger.Error(ctx, "Failed to get config",
				observability.String("key", key),
				observability.Error(err))
		}
		return "", fmt.Errorf("failed to get config: %w", err)
	}

	if pair == nil {
		return "", fmt.Errorf("config key not found: %s", key)
	}

	return string(pair.Value), nil
}

// SetConfig 设置配置到Consul KV
func (c *DefaultConsulClient) SetConfig(ctx context.Context, key, value string) error {
	pair := &api.KVPair{
		Key:   key,
		Value: []byte(value),
	}

	_, err := c.client.KV().Put(pair, nil)
	if err != nil {
		if c.logger != nil {
			c.logger.Error(ctx, "Failed to set config",
				observability.String("key", key),
				observability.Error(err))
		}
		return fmt.Errorf("failed to set config: %w", err)
	}

	if c.logger != nil {
		c.logger.Debug(ctx, "Config set successfully",
			observability.String("key", key))
	}

	return nil
}

// 中间件实现

// ServiceDiscoveryMiddleware 服务发现中间件
func ServiceDiscoveryMiddleware(consulClient ConsulClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 将consul客户端存储在context中，供后续使用
		c.Set("consul_client", consulClient)
		c.Next()
	}
}

// HealthCheckMiddleware 健康检查中间件
func HealthCheckMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" {
			response := map[string]any{
				"service":   serviceName,
				"status":    "healthy",
				"timestamp": time.Now().Format(time.RFC3339),
			}
			c.JSON(http.StatusOK, response)
			c.Abort()
			return
		}
		c.Next()
	}
}

// RegisterService 注册服务的便利函数
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
