package service

import (
	"fmt"
	"mocks3/shared/observability/config"
	"mocks3/shared/utils"
)

// Config Metadata Service配置
type Config struct {
	Service struct {
		Name string `yaml:"name"`
		Port int    `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"service"`

	Consul struct {
		Address string `yaml:"address"`
	} `yaml:"consul"`

	Database struct {
		Driver   string `yaml:"driver"`
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Name     string `yaml:"name"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"database"`

	Redis struct {
		Address  string `yaml:"address"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`

	Observability config.ObservabilityConfig `yaml:"observability"`
}

// LoadConfig 加载配置
func LoadConfig(configPath string) (*Config, error) {
	config := &Config{}
	err := utils.LoadConfig(configPath, config)
	if err != nil {
		return nil, err
	}

	// 设置默认值
	if config.Service.Host == "" {
		config.Service.Host = "0.0.0.0"
	}
	if config.Service.Port == 0 {
		config.Service.Port = 8081
	}

	return config, nil
}

// GetDSN 获取数据库连接字符串
func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Database.Host, c.Database.Port, c.Database.Username, c.Database.Password, c.Database.Name)
}

// GetRedisAddr 获取Redis地址
func (c *Config) GetRedisAddr() string {
	return c.Redis.Address
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
