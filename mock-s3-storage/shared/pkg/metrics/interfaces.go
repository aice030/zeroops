package metrics

import (
	"context"
	"net/http"
	"time"
)

// Metrics 核心指标接口 - 简化的统一接口
type Metrics interface {
	// 计数器操作
	IncCounter(name string, labels ...string)
	AddCounter(name string, value float64, labels ...string)

	// 测量器操作  
	SetGauge(name string, value float64, labels ...string)
	IncGauge(name string, labels ...string)
	DecGauge(name string, labels ...string)

	// 直方图操作
	ObserveHistogram(name string, value float64, labels ...string)
	ObserveDuration(name string, duration time.Duration, labels ...string)

	// Timer操作
	StartTimer(name string, labels ...string) Timer

	// HTTP指标（预定义高频指标）
	RecordHTTPRequest(method, endpoint, status string, duration time.Duration)
	RecordError(errorType, operation string)

	// 服务相关
	HTTPHandler() http.Handler
	Middleware() func(http.Handler) http.Handler
	Shutdown(ctx context.Context) error
}

// Timer 简化的计时器接口
type Timer interface {
	ObserveDuration() time.Duration
}

// Config 指标配置
type Config struct {
	ServiceName string            `json:"service_name"`
	ServiceVer  string            `json:"service_version"`
	Namespace   string            `json:"namespace"`
	Labels      map[string]string `json:"labels"`
}

// Stats 简化的统计信息
type Stats struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
	Type  string  `json:"type"` // counter, gauge, histogram
}