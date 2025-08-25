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

// Config 完整配置结构
type Config struct {
	Service  ServiceConfig  `yaml:"service"`
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
