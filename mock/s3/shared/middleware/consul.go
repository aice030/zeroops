package middleware

import (
	"context"
	"fmt"
	"log"
	"mocks3/shared/models"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"
)

// ConsulManager Consul管理器
type ConsulManager struct {
	client      *api.Client
	serviceName string
	serviceID   string
	servicePort int
	healthCheck *api.AgentServiceCheck
}

// ConsulConfig Consul配置
type ConsulConfig struct {
	Address     string
	ServiceName string
	ServicePort int
	ServiceHost string // 服务主机地址
	HealthPath  string
	Tags        []string
	Metadata    map[string]string
}

// NewConsulManager 创建Consul管理器
func NewConsulManager(config *ConsulConfig) (*ConsulManager, error) {
	consulConfig := api.DefaultConfig()
	consulConfig.Address = config.Address

	client, err := api.NewClient(consulConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	// 生成唯一的服务ID
	hostname, _ := os.Hostname()
	serviceID := fmt.Sprintf("%s-%s-%d", config.ServiceName, hostname, config.ServicePort)

	// 获取服务主机地址
	serviceHost := config.ServiceHost
	if serviceHost == "" {
		// 在容器环境中使用主机名
		serviceHost = hostname
	}

	// 创建健康检查
	healthCheck := &api.AgentServiceCheck{
		HTTP:                           fmt.Sprintf("http://%s:%d%s", serviceHost, config.ServicePort, config.HealthPath),
		Interval:                       "10s",
		Timeout:                        "5s",
		DeregisterCriticalServiceAfter: "30s",
	}

	return &ConsulManager{
		client:      client,
		serviceName: config.ServiceName,
		serviceID:   serviceID,
		servicePort: config.ServicePort,
		healthCheck: healthCheck,
	}, nil
}

// RegisterService 注册服务
func (cm *ConsulManager) RegisterService(ctx context.Context, config *ConsulConfig) error {
	// 获取服务地址
	serviceAddress := config.ServiceHost
	if serviceAddress == "" {
		// 在容器环境中使用主机名
		hostname, _ := os.Hostname()
		serviceAddress = hostname
	}

	service := &api.AgentServiceRegistration{
		ID:      cm.serviceID,
		Name:    cm.serviceName,
		Port:    cm.servicePort,
		Address: serviceAddress,
		Tags:    config.Tags,
		Meta:    config.Metadata,
		Check:   cm.healthCheck,
	}

	err := cm.client.Agent().ServiceRegister(service)
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	log.Printf("Service registered: %s (ID: %s)", cm.serviceName, cm.serviceID)
	return nil
}

// DeregisterService 注销服务
func (cm *ConsulManager) DeregisterService(ctx context.Context) error {
	err := cm.client.Agent().ServiceDeregister(cm.serviceID)
	if err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}

	log.Printf("Service deregistered: %s (ID: %s)", cm.serviceName, cm.serviceID)
	return nil
}

// DiscoverServices 发现服务
func (cm *ConsulManager) DiscoverServices(ctx context.Context, serviceName string) ([]*models.ServiceInfo, error) {
	services, _, err := cm.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to discover services: %w", err)
	}

	var serviceInfos []*models.ServiceInfo
	for _, service := range services {
		serviceInfo := &models.ServiceInfo{
			ID:       service.Service.ID,
			Name:     service.Service.Service,
			Address:  service.Service.Address,
			Port:     service.Service.Port,
			Tags:     service.Service.Tags,
			Metadata: service.Service.Meta,
			Health: models.ServiceHealth{
				Status: models.HealthStatusHealthy, // 这里只返回健康的服务
			},
		}
		serviceInfos = append(serviceInfos, serviceInfo)
	}

	return serviceInfos, nil
}

// GetConfig 获取配置
func (cm *ConsulManager) GetConfig(ctx context.Context, key string) (string, error) {
	kv, _, err := cm.client.KV().Get(key, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get config: %w", err)
	}

	if kv == nil {
		return "", fmt.Errorf("config key not found: %s", key)
	}

	return string(kv.Value), nil
}

// SetConfig 设置配置
func (cm *ConsulManager) SetConfig(ctx context.Context, key, value string) error {
	kv := &api.KVPair{
		Key:   key,
		Value: []byte(value),
	}

	_, err := cm.client.KV().Put(kv, nil)
	if err != nil {
		return fmt.Errorf("failed to set config: %w", err)
	}

	return nil
}

// WatchConfig 监听配置变化
func (cm *ConsulManager) WatchConfig(ctx context.Context, key string) (<-chan string, error) {
	ch := make(chan string, 1)

	go func() {
		defer close(ch)

		var lastIndex uint64
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			queryOptions := &api.QueryOptions{
				WaitIndex: lastIndex,
				WaitTime:  30 * time.Second,
			}

			kv, meta, err := cm.client.KV().Get(key, queryOptions)
			if err != nil {
				log.Printf("Error watching config %s: %v", key, err)
				time.Sleep(5 * time.Second)
				continue
			}

			if meta.LastIndex == lastIndex {
				continue
			}

			lastIndex = meta.LastIndex

			if kv != nil {
				select {
				case ch <- string(kv.Value):
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return ch, nil
}

// GetServiceHealth 获取服务健康状态
func (cm *ConsulManager) GetServiceHealth(ctx context.Context, serviceID string) (bool, error) {
	checks, _, err := cm.client.Health().Checks(cm.serviceName, nil)
	if err != nil {
		return false, fmt.Errorf("failed to get service health: %w", err)
	}

	for _, check := range checks {
		if check.ServiceID == serviceID {
			return check.Status == api.HealthPassing, nil
		}
	}

	return false, fmt.Errorf("service not found: %s", serviceID)
}

// SetServiceHealth 设置服务健康状态
func (cm *ConsulManager) SetServiceHealth(ctx context.Context, serviceID string, healthy bool) error {
	status := api.HealthPassing
	if !healthy {
		status = api.HealthCritical
	}

	err := cm.client.Agent().UpdateTTL(fmt.Sprintf("service:%s", serviceID), "Manual health update", status)
	if err != nil {
		return fmt.Errorf("failed to set service health: %w", err)
	}

	return nil
}

// LoadServiceConfigFromConsul 从 Consul KV 加载服务配置
func LoadServiceConfigFromConsul(consulAddr, serviceName string) (*ConsulConfig, error) {
	// 创建 Consul 客户端
	consulConfig := api.DefaultConfig()
	consulConfig.Address = consulAddr
	client, err := api.NewClient(consulConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	// 配置默认值
	config := &ConsulConfig{
		Address:     consulAddr,
		ServiceName: serviceName,
		ServicePort: 8080,
		ServiceHost: "",
		HealthPath:  "/health",
		Tags:        []string{"api", "http"},
		Metadata:    make(map[string]string),
	}

	// 从 Consul KV 加载配置
	ctx := context.Background()

	// 加载服务端口
	if port, err := getConsulValue(client, ctx, fmt.Sprintf("services/%s/port", serviceName)); err == nil && port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.ServicePort = p
		}
	}

	// 加载服务主机
	if host, err := getConsulValue(client, ctx, fmt.Sprintf("services/%s/host", serviceName)); err == nil && host != "" {
		config.ServiceHost = host
	}

	// 加载健康检查路径
	if health, err := getConsulValue(client, ctx, fmt.Sprintf("services/%s/health_path", serviceName)); err == nil && health != "" {
		config.HealthPath = health
	}

	// 加载标签
	if tags, err := getConsulValue(client, ctx, fmt.Sprintf("services/%s/tags", serviceName)); err == nil && tags != "" {
		config.Tags = strings.Split(tags, ",")
	}

	// 加载元数据
	if version, err := getConsulValue(client, ctx, fmt.Sprintf("services/%s/version", serviceName)); err == nil && version != "" {
		config.Metadata["version"] = version
	} else {
		config.Metadata["version"] = "1.0.0"
	}

	if env, err := getConsulValue(client, ctx, fmt.Sprintf("services/%s/environment", serviceName)); err == nil && env != "" {
		config.Metadata["environment"] = env
	} else {
		config.Metadata["environment"] = "development"
	}

	return config, nil
}

// getConsulValue 从 Consul KV 获取值
func getConsulValue(client *api.Client, ctx context.Context, key string) (string, error) {
	kv, _, err := client.KV().Get(key, nil)
	if err != nil {
		return "", err
	}
	if kv == nil {
		return "", fmt.Errorf("key not found: %s", key)
	}
	return string(kv.Value), nil
}

// NewDefaultConsulManager 创建默认的Consul管理器
func NewDefaultConsulManager(serviceName string) (*ConsulManager, error) {
	// 只从环境变量获取 Consul 地址，其他配置都从 Consul KV 加载
	consulAddr := getEnv("CONSUL_ADDR", "consul:8500")

	config, err := LoadServiceConfigFromConsul(consulAddr, serviceName)
	if err != nil {
		// 如果无法从 Consul 加载配置，使用默认配置
		log.Printf("Failed to load config from Consul, using defaults: %v", err)
		config = &ConsulConfig{
			Address:     consulAddr,
			ServiceName: serviceName,
			ServicePort: 8080,
			ServiceHost: "",
			HealthPath:  "/health",
			Tags:        []string{"api", "http"},
			Metadata: map[string]string{
				"version":     "1.0.0",
				"environment": "development",
			},
		}
	}

	return NewConsulManager(config)
}

// InitializeServiceConfig 初始化服务配置到 Consul KV
func InitializeServiceConfig(consulAddr, serviceName string, defaultPort int) error {
	consulConfig := api.DefaultConfig()
	consulConfig.Address = consulAddr
	client, err := api.NewClient(consulConfig)
	if err != nil {
		return fmt.Errorf("failed to create consul client: %w", err)
	}

	ctx := context.Background()
	servicePrefix := fmt.Sprintf("services/%s", serviceName)

	// 默认配置
	defaultConfigs := map[string]string{
		"port":        strconv.Itoa(defaultPort),
		"host":        "", // 空表示使用主机名
		"health_path": "/health",
		"tags":        "api,http",
		"version":     "1.0.0",
		"environment": "development",
	}

	// 只设置不存在的配置
	for key, value := range defaultConfigs {
		fullKey := fmt.Sprintf("%s/%s", servicePrefix, key)
		if _, err := getConsulValue(client, ctx, fullKey); err != nil {
			// 配置不存在，设置默认值
			kv := &api.KVPair{
				Key:   fullKey,
				Value: []byte(value),
			}
			if _, err := client.KV().Put(kv, nil); err != nil {
				return fmt.Errorf("failed to set default config %s: %w", fullKey, err)
			}
			log.Printf("Set default config: %s = %s", fullKey, value)
		}
	}

	return nil
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
