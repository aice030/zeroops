package service

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// MockErrorConfig Mock Error Service配置
type MockErrorConfig struct {
	Service ServiceConfig `yaml:"service"`
	Storage StorageConfig `yaml:"storage"`
	Consul  ConsulConfig  `yaml:"consul"`
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	Name string `yaml:"name"`
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	DataDir string `yaml:"data_dir"`
}

// ConsulConfig Consul配置
type ConsulConfig struct {
	Address string `yaml:"address"`
}

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*MockErrorConfig, error) {
	// 确保配置文件路径是绝对路径
	if !filepath.IsAbs(configPath) {
		// 假设配置文件在相对于可执行文件的config目录下
		executableDir, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get executable directory: %w", err)
		}
		configPath = filepath.Join(executableDir, configPath)
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// 解析YAML
	var config MockErrorConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 设置默认值
	if config.Service.Name == "" {
		config.Service.Name = "mock-error-service"
	}
	if config.Service.Host == "" {
		config.Service.Host = "0.0.0.0"
	}
	if config.Service.Port == 0 {
		config.Service.Port = 8085
	}
	if config.Storage.DataDir == "" {
		config.Storage.DataDir = "./data"
	}
	if config.Consul.Address == "" {
		config.Consul.Address = "localhost:8500"
	}

	return &config, nil
}

// GetServiceName 实现server.ServiceConfig接口
func (c *MockErrorConfig) GetServiceName() string {
	return c.Service.Name
}

// GetHost 实现server.ServiceConfig接口
func (c *MockErrorConfig) GetHost() string {
	return c.Service.Host
}

// GetPort 实现server.ServiceConfig接口
func (c *MockErrorConfig) GetPort() int {
	return c.Service.Port
}

// GetConsulAddress 实现server.ConsulServiceConfig接口
func (c *MockErrorConfig) GetConsulAddress() string {
	return c.Consul.Address
}
