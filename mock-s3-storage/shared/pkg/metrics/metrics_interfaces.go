package metrics

import (
	"context"
	"net/http"
	"time"

	"shared/pkg/config"
)

type CounterCollector interface {
	IncCounter(name string, labels ...string)
	AddCounter(name string, value float64, labels ...string)
}

type GaugeCollector interface {
	SetGauge(name string, value float64, labels ...string)
	IncGauge(name string, labels ...string)
	DecGauge(name string, labels ...string)
}

type HistogramCollector interface {
	ObserveHistogram(name string, value float64, labels ...string)
	StartTimer(name string, labels ...string) Timer
}

// Timer 计时器接口
type Timer interface {
	ObserveDuration() time.Duration
	Cancel()      // 取消计时器但不记录
	Reset() Timer // 重置计时器
}

type TimerCollector interface {
	StartTimer(name string, labels ...string) Timer
}

type HTTPCollector interface {
	RecordHTTPRequest(method, endpoint, status string, duration time.Duration)
	RecordError(errorType, operation string)
	Middleware() func(http.Handler) http.Handler
}

// ConfigManager 配置管理接口
type ConfigManager interface {
	GetConfig() config.MetricsConfig
	IsEnabled() bool
	UpdateConfig(config.MetricsConfig) error
}

// MetricsServer 服务器接口
type MetricsServer interface {
	HTTPHandler() http.Handler
	Start() error
	Stop(ctx context.Context) error
}

// HealthChecker 健康检查接口
type HealthChecker interface {
	HealthCheck() error
	IsHealthy() bool
}

// LifecycleManager 生命周期管理接口
type LifecycleManager interface {
	Initialize() error
	Shutdown(ctx context.Context) error
	Reset() error
}

// Stats 指标统计信息
type Stats struct {
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Type      string            `json:"type"` // counter, gauge, histogram
	Labels    map[string]string `json:"labels,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// MetricsSnapshot 指标快照
type MetricsSnapshot struct {
	ServiceName string            `json:"service_name"`
	Timestamp   time.Time         `json:"timestamp"`
	Stats       []Stats           `json:"stats"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// MetricsReader 指标查询接口
type MetricsReader interface {
	GetStats() []Stats
	GetMetricValue(name string, labels ...string) (float64, bool)
	GetSnapshot() MetricsSnapshot
}

// MetricsCollector 综合指标收集接口
type MetricsCollector interface {
	CounterCollector
	GaugeCollector
	HistogramCollector
	TimerCollector
	HTTPCollector
	MetricsReader
	HealthChecker
	LifecycleManager
}
