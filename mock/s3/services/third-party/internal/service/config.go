package service

import (
	"mocks3/shared/observability/config"
	"mocks3/shared/utils"
)

// Config Third-Party Service配置
type Config struct {
	Service struct {
		Name string `yaml:"name"`
		Port int    `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"service"`

	Consul struct {
		Address string `yaml:"address"`
	} `yaml:"consul"`

	DataSources []DataSource `yaml:"data_sources"`

	Mock struct {
		Enabled            bool    `yaml:"enabled"`
		DefaultContentType string  `yaml:"default_content_type"`
		LatencyMs          int     `yaml:"latency_ms"`
		SuccessRate        float64 `yaml:"success_rate"`
	} `yaml:"mock"`

	Observability config.ObservabilityConfig `yaml:"observability"`
}

// DataSource 第三方数据源配置
type DataSource struct {
	Name       string `yaml:"name"`
	URL        string `yaml:"url"`
	TimeoutMs  int    `yaml:"timeout_ms"`
	RetryCount int    `yaml:"retry_count"`
	Enabled    bool   `yaml:"enabled"`
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
		config.Service.Port = 8084
	}
	if config.Mock.DefaultContentType == "" {
		config.Mock.DefaultContentType = "application/octet-stream"
	}
	if config.Mock.LatencyMs == 0 {
		config.Mock.LatencyMs = 100
	}
	if config.Mock.SuccessRate == 0 {
		config.Mock.SuccessRate = 0.9
	}

	return config, nil
}

// GetEnabledDataSources 获取启用的数据源
func (c *Config) GetEnabledDataSources() []DataSource {
	var enabled []DataSource
	for _, ds := range c.DataSources {
		if ds.Enabled {
			enabled = append(enabled, ds)
		}
	}
	return enabled
}
