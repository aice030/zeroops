package models

import (
	"time"
)

// ServiceInfo 服务信息
type ServiceInfo struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Address      string            `json:"address"`
	Port         int               `json:"port"`
	Tags         []string          `json:"tags,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	Health       ServiceHealth     `json:"health"`
	Weight       int               `json:"weight"` // 负载均衡权重
	Version      string            `json:"version,omitempty"`
	RegisteredAt time.Time         `json:"registered_at"`
	LastSeen     time.Time         `json:"last_seen"`
}

// ServiceHealth 服务健康状态
type ServiceHealth struct {
	Status    HealthStatus  `json:"status"`
	CheckURL  string        `json:"check_url,omitempty"`
	Interval  time.Duration `json:"interval,omitempty"`
	Timeout   time.Duration `json:"timeout,omitempty"`
	LastCheck time.Time     `json:"last_check"`
	Message   string        `json:"message,omitempty"`
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
	Success   bool      `json:"success"`
	Data      any       `json:"data,omitempty"`
	Error     *APIError `json:"error,omitempty"`
	Message   string    `json:"message,omitempty"`
	RequestID string    `json:"request_id,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// APIError API 错误
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// HealthCheckResponse 健康检查响应
type HealthCheckResponse struct {
	Status    HealthStatus           `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version,omitempty"`
	Uptime    time.Duration          `json:"uptime"`
	Checks    map[string]CheckResult `json:"checks,omitempty"`
}

// CheckResult 检查结果
type CheckResult struct {
	Status  HealthStatus  `json:"status"`
	Message string        `json:"message,omitempty"`
	Latency time.Duration `json:"latency,omitempty"`
}

// ConfigItem 配置项
type ConfigItem struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Version   int64     `json:"version"`
	UpdatedBy string    `json:"updated_by,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RateLimit 限流配置（简化版）
type RateLimit struct {
	RequestsPerSecond int           `json:"requests_per_second"`
	BurstSize         int           `json:"burst_size"`
	WindowDuration    time.Duration `json:"window_duration"`
}
