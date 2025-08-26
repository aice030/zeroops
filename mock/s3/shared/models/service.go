package models

import (
	"time"
)

// ServiceInfo 服务信息（简化版）
type ServiceInfo struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	Address  string       `json:"address"`
	Port     int          `json:"port"`
	Tags     []string     `json:"tags,omitempty"`
	Health   HealthStatus `json:"health"`
	Weight   int          `json:"weight"`
	Version  string       `json:"version,omitempty"`
	LastSeen time.Time    `json:"last_seen"`
}

// HealthStatus 健康状态
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
	HealthStatusCritical  HealthStatus = "critical"
)

// APIResponse 通用 API 响应
type APIResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

// HealthCheckResponse 健康检查响应
type HealthCheckResponse struct {
	Status    HealthStatus  `json:"status"`
	Timestamp time.Time     `json:"timestamp"`
	Uptime    time.Duration `json:"uptime"`
	Message   string        `json:"message,omitempty"`
}

// ConfigItem 配置项
type ConfigItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// RateLimit 限流配置
type RateLimit struct {
	RequestsPerSecond int           `json:"requests_per_second"`
	BurstSize         int           `json:"burst_size"`
	WindowDuration    time.Duration `json:"window_duration"`
}
