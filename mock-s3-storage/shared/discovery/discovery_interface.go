package discovery

import (
	"context"
	"shared/config"
)

// ServiceInfo 服务信息
type ServiceInfo struct {
	ID      string `json:"id"`      // 服务实例ID
	Name    string `json:"name"`    // 服务名称
	Address string `json:"address"` // 服务地址
	Port    int    `json:"port"`    // 服务端口
	Health  bool   `json:"health"`  // 健康状态（简化为bool）
}

// ServiceDiscovery 服务发现接口
type ServiceDiscovery interface {
	// Register 注册服务
	Register(ctx context.Context, service *ServiceInfo) error

	// Deregister 注销服务
	Deregister(ctx context.Context, serviceID string) error

	// Discover 发现服务实例
	Discover(ctx context.Context, serviceName string) ([]*ServiceInfo, error)

	// Close 关闭服务发现客户端
	Close() error
}

// ServiceRegistry 服务注册器
type ServiceRegistry interface {
	// RegisterSelf 注册当前服务
	RegisterSelf(ctx context.Context, name, address string, port int) error

	// DeregisterSelf 注销当前服务
	DeregisterSelf(ctx context.Context) error

	// HealthCheck 更新健康状态
	HealthCheck(ctx context.Context) error

	// GetServiceID 获取当前服务ID
	GetServiceID() string
}

// NewConsulDiscovery 创建Consul服务发现
func NewConsulDiscovery(config config.ConsulConfig) (ServiceDiscovery, error) {
	// TODO: 实现Consul服务发现
	return nil, nil
}

// NewRegistry 创建服务注册器
func NewRegistry(discovery ServiceDiscovery, serviceInfo ServiceInfo) ServiceRegistry {
	// TODO: 实现服务注册器
	return nil
}

// SimpleLoadBalancer 简单的负载均衡（轮询）
func SimpleLoadBalancer(services []*ServiceInfo) *ServiceInfo {
	if len(services) == 0 {
		return nil
	}
	// TODO: 实现简单轮询逻辑
	return services[0]
}
