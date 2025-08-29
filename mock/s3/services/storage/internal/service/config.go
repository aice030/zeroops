package service

import (
	"fmt"
	"mocks3/shared/utils"
	"time"
)

// StorageNode 存储节点配置
type StorageNode struct {
	ID   string `yaml:"id"`
	Path string `yaml:"path"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	RootPath string        `yaml:"root_path"`
	Nodes    []StorageNode `yaml:"nodes"`
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	Name string `yaml:"name"`
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// ExternalService 外部服务配置
type ExternalService struct {
	URL     string `yaml:"url"`
	Timeout string `yaml:"timeout"`
}

// ServicesConfig 外部服务配置
type ServicesConfig struct {
	Metadata   ExternalService `yaml:"metadata"`
	Queue      ExternalService `yaml:"queue"`
	ThirdParty ExternalService `yaml:"third_party"`
}

// ConsulConfig Consul配置
type ConsulConfig struct {
	Address string `yaml:"address"`
}

// Config 完整配置结构
type Config struct {
	Service  ServiceConfig  `yaml:"service"`
	Consul   ConsulConfig   `yaml:"consul"`
	Storage  StorageConfig  `yaml:"storage"`
	Services ServicesConfig `yaml:"services"`
}

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	var config Config
	if err := utils.LoadConfig(configPath, &config); err != nil {
		return nil, err
	}

	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	// 设置默认值
	if config.Consul.Address == "" {
		config.Consul.Address = "localhost:8500"
	}

	return &config, nil
}

// validate 验证配置
func (c *Config) validate() error {
	if c.Service.Name == "" {
		return fmt.Errorf("service name is required")
	}
	if c.Service.Port <= 0 {
		return fmt.Errorf("service port must be positive")
	}
	if len(c.Storage.Nodes) == 0 {
		return fmt.Errorf("at least one storage node is required")
	}
	return nil
}

// GetMetadataTimeout 获取元数据服务超时时间
func (c *Config) GetMetadataTimeout() time.Duration {
	if duration, err := time.ParseDuration(c.Services.Metadata.Timeout); err == nil {
		return duration
	}
	return 30 * time.Second
}

// GetQueueTimeout 获取队列服务超时时间
func (c *Config) GetQueueTimeout() time.Duration {
	if duration, err := time.ParseDuration(c.Services.Queue.Timeout); err == nil {
		return duration
	}
	return 10 * time.Second
}

// GetThirdPartyTimeout 获取第三方服务超时时间
func (c *Config) GetThirdPartyTimeout() time.Duration {
	if duration, err := time.ParseDuration(c.Services.ThirdParty.Timeout); err == nil {
		return duration
	}
	return 15 * time.Second
}

// GetServiceName 实现server.ServiceConfig接口
func (c *Config) GetServiceName() string {
	return c.Service.Name
}

// GetHost 实现server.ServiceConfig接口
func (c *Config) GetHost() string {
	return c.Service.Host
}

// GetPort 实现server.ServiceConfig接口
func (c *Config) GetPort() int {
	return c.Service.Port
}

// GetConsulAddress 实现server.ConsulServiceConfig接口
func (c *Config) GetConsulAddress() string {
	return c.Consul.Address
}
