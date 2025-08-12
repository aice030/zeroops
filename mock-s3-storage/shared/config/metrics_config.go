package config

import "shared/errors"

// MetricsConfig 指标配置
type MetricsConfig struct {
	ServiceName string            `json:"service_name" yaml:"service_name"`
	ServiceVer  string            `json:"service_version" yaml:"service_version"`
	Namespace   string            `json:"namespace" yaml:"namespace"`
	Labels      map[string]string `json:"labels" yaml:"labels"`
	Enabled     bool              `json:"enabled" yaml:"enabled"`
	Port        int               `json:"port" yaml:"port"`
	Path        string            `json:"path" yaml:"path"`
}

// DefaultMetricsConfig 返回默认的指标配置
func DefaultMetricsConfig() MetricsConfig {
	return MetricsConfig{
		ServiceName: "unknown",
		ServiceVer:  "unknown",
		Namespace:   "mock_s3",
		Labels:      make(map[string]string),
		Enabled:     true,
		Port:        9090,
		Path:        "/metrics",
	}
}

// NewMetricsConfig 创建新的指标配置
func NewMetricsConfig(serviceName, version string) MetricsConfig {
	config := DefaultMetricsConfig()
	config.ServiceName = serviceName
	config.ServiceVer = version
	return config
}

// SetNamespace 设置命名空间
func (c MetricsConfig) SetNamespace(namespace string) MetricsConfig {
	c.Namespace = namespace
	return c
}

// SetLabels 设置标签
func (c MetricsConfig) SetLabels(labels map[string]string) MetricsConfig {
	if c.Labels == nil {
		c.Labels = make(map[string]string)
	}
	for k, v := range labels {
		c.Labels[k] = v
	}
	return c
}

// SetLabel 添加单个标签
func (c MetricsConfig) SetLabel(key, value string) MetricsConfig {
	if c.Labels == nil {
		c.Labels = make(map[string]string)
	}
	c.Labels[key] = value
	return c
}

// SetPort 设置端口
func (c MetricsConfig) SetPort(port int) MetricsConfig {
	c.Port = port
	return c
}

// SetPath 设置路径
func (c MetricsConfig) SetPath(path string) MetricsConfig {
	c.Path = path
	return c
}

// Disable 禁用指标
func (c MetricsConfig) Disable() MetricsConfig {
	c.Enabled = false
	return c
}

// Validate 验证配置
func (c MetricsConfig) Validate() error {
	validationErr := errors.NewConfigValidationError("config", "validate_metrics_config")

	if c.ServiceName == "" {
		validationErr.AddFieldError("service_name", "service name cannot be empty")
	}
	if c.Namespace == "" {
		validationErr.AddFieldError("namespace", "namespace cannot be empty")
	}
	if c.Port <= 0 || c.Port > 65535 {
		validationErr.AddFieldError("port", "port must be between 1 and 65535", c.Port)
	}
	if c.Path == "" {
		validationErr.AddFieldError("path", "path cannot be empty")
	}

	if validationErr.HasErrors() {
		return validationErr
	}
	return nil
}
