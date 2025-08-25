package models

import (
	"time"
)

// ErrorRule 错误注入规则
type ErrorRule struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Service   string `json:"service"`   // 目标服务
	Operation string `json:"operation"` // 目标操作
	Enabled   bool   `json:"enabled"`

	// 触发条件
	Conditions map[string]any `json:"conditions"`

	// 错误动作
	Action map[string]any `json:"action"`

	// 时间和计数
	StartTime   *time.Time     `json:"start_time,omitempty"`
	Duration    *time.Duration `json:"duration,omitempty"`
	MaxTriggers int            `json:"max_triggers"`
	Triggered   int            `json:"triggered"`
	CreatedAt   time.Time      `json:"created_at"`
}

// 错误类型
const (
	// 错误动作类型
	ErrorHTTP       = "http_error"
	ErrorNetwork    = "network_error"
	ErrorTimeout    = "timeout"
	ErrorDelay      = "delay"
	ErrorCorruption = "corruption"
	ErrorDisconnect = "disconnect"
	ErrorDatabase   = "database_error"
	ErrorStorage    = "storage_error"
	ErrorMetric     = "metric_anomaly"

	// 指标异常类型
	MetricCPU     = "cpu_spike"
	MetricMemory  = "memory_leak"
	MetricDisk    = "disk_full"
	MetricNetwork = "network_flood"
	MetricMachine = "machine_down"
)

// 使用示例：
//
// HTTP 错误规则:
// {
//   "action": {
//     "type": "http_error",
//     "http_code": 500,
//     "message": "Internal Server Error"
//   }
// }
//
// 指标异常规则:
// {
//   "action": {
//     "type": "metric_anomaly",
//     "metric_name": "system_cpu_usage_percent",
//     "anomaly_type": "cpu_spike",
//     "target_value": 95.0,
//     "duration": 300
//   }
// }
