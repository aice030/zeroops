package service

import (
	"fmt"
	"mocks3/services/queue/internal/worker"
	"mocks3/shared/observability/config"
	"mocks3/shared/utils"
	"time"
)

// ServiceConfig 服务配置
type ServiceConfig struct {
	Name string `yaml:"name"`
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	URL      string `yaml:"url"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// ConsulConfig Consul配置
type ConsulConfig struct {
	Address string `yaml:"address"`
}

// Config Queue Service配置
type Config struct {
	Service       ServiceConfig              `yaml:"service"`
	Consul        ConsulConfig               `yaml:"consul"`
	Redis         RedisConfig                `yaml:"redis"`
	Worker        worker.WorkerConfig        `yaml:"worker"`
	Observability config.ObservabilityConfig `yaml:"observability"`
}

// LoadConfig 加载配置
func LoadConfig(configPath string) (*Config, error) {
	config := &Config{}
	if err := utils.LoadConfig(configPath, config); err != nil {
		return nil, err
	}

	// 设置默认值
	if config.Service.Host == "" {
		config.Service.Host = "0.0.0.0"
	}
	if config.Service.Port == 0 {
		config.Service.Port = 8083
	}
	if config.Redis.URL == "" {
		config.Redis.URL = "redis://localhost:6379"
	}
	if config.Worker.WorkerCount <= 0 {
		config.Worker.WorkerCount = 3
	}
	if config.Worker.PollInterval <= 0 {
		config.Worker.PollInterval = 5 * time.Second
	}
	if config.Consul.Address == "" {
		config.Consul.Address = "localhost:8500"
	}

	return config, nil
}

// GetRedisURL 获取完整的Redis连接URL
func (c *Config) GetRedisURL() string {
	if c.Redis.Password != "" {
		return fmt.Sprintf("redis://:%s@%s/%d", c.Redis.Password, c.Redis.URL[8:], c.Redis.DB)
	}
	if c.Redis.DB != 0 {
		return fmt.Sprintf("%s/%d", c.Redis.URL, c.Redis.DB)
	}
	return c.Redis.URL
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

// GetConsulAddress 获取Consul地址
func (c *Config) GetConsulAddress() string {
	return c.Consul.Address
}
