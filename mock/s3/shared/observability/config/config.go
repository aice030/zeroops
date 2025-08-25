package config

import (
	"fmt"
	"mocks3/shared/utils"
	"os"
	"time"
)

// ObservabilityConfig 可观测性配置
type ObservabilityConfig struct {
	ServiceName    string        `yaml:"service_name"`
	ServiceVersion string        `yaml:"service_version"`
	Environment    string        `yaml:"environment"`
	OTLPEndpoint   string        `yaml:"otlp_endpoint"`
	LogLevel       string        `yaml:"log_level"`
	SamplingRatio  float64       `yaml:"sampling_ratio"`
	ExportInterval time.Duration `yaml:"export_interval"`
}

// LoadObservabilityConfig 从YAML配置文件加载可观测性配置
func LoadObservabilityConfig(serviceName, configPath string) (*ObservabilityConfig, error) {
	// 默认配置
	config := &ObservabilityConfig{
		ServiceName:    serviceName,
		ServiceVersion: "1.0.0",
		Environment:    "development",
		OTLPEndpoint:   "localhost:4318",
		LogLevel:       "info",
		SamplingRatio:  1.0,
		ExportInterval: 30 * time.Second,
	}

	// 如果配置文件存在，则读取YAML配置
	if configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			if err := utils.LoadConfigFromYAML(configPath, config); err != nil {
				return nil, err
			}
		}
	}

	// 设置服务名称（如果配置文件中没有指定）
	if config.ServiceName == "" {
		config.ServiceName = serviceName
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return config, nil
}

// Validate 验证可观测性配置
func (c *ObservabilityConfig) Validate() error {
	if c.ServiceName == "" {
		return fmt.Errorf("service_name is required")
	}
	if c.OTLPEndpoint == "" {
		return fmt.Errorf("otlp_endpoint is required")
	}
	if c.SamplingRatio < 0 || c.SamplingRatio > 1 {
		return fmt.Errorf("sampling_ratio must be between 0 and 1")
	}
	if c.ExportInterval <= 0 {
		return fmt.Errorf("export_interval must be positive")
	}

	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[c.LogLevel] {
		return fmt.Errorf("invalid log_level: %s", c.LogLevel)
	}

	return nil
}
