package metrics

import (
	"context"
	"net/http"
	"time"
)

// 基础指标接口定义

// Counter 计数器接口 - 只能递增的指标
type Counter interface {
	Inc()                                          // 增加1
	Add(value float64)                             // 增加指定值
	Get() float64                                  // 获取当前值
	Reset()                                        // 重置为0
	WithLabels(labels map[string]string) Counter   // 添加标签
}

// Gauge 测量器接口 - 可增可减的指标
type Gauge interface {
	Inc()                                        // 增加1
	Dec()                                        // 减少1
	Add(value float64)                           // 增加指定值
	Sub(value float64)                           // 减少指定值
	Set(value float64)                           // 设置值
	Get() float64                                // 获取当前值
	WithLabels(labels map[string]string) Gauge   // 添加标签
}

// Histogram 直方图接口 - 记录分布情况
type Histogram interface {
	Observe(value float64)                             // 记录观测值
	Count() uint64                                     // 获取观测次数
	Sum() float64                                      // 获取观测值总和
	Quantile(q float64) float64                        // 获取分位数
	Reset()                                            // 重置直方图
	WithLabels(labels map[string]string) Histogram     // 添加标签
}

// Timer 计时器接口 - 自动记录耗时
type Timer interface {
	ObserveDuration() time.Duration    // 记录从创建到现在的耗时
	Reset()                           // 重新开始计时
}

// MetricsProvider 指标提供者接口 - 统一的指标创建和管理
type MetricsProvider interface {
	// 基础指标创建
	NewCounter(name, help string, labelNames []string) (Counter, error)
	NewGauge(name, help string, labelNames []string) (Gauge, error)
	NewHistogram(name, help string, buckets []float64, labelNames []string) (Histogram, error)
	
	// 预定义指标访问
	HTTPRequestsTotal() Counter
	HTTPRequestDuration() Histogram
	ErrorsTotal() Counter
	ActiveRequests() Gauge
	
	// 指标操作
	GetAllMetricNames() []string
	Reset()
	
	// 服务器相关
	HTTPHandler() http.Handler
	StartMetricsServer(addr string) error
	Shutdown(ctx context.Context) error
}

// PerformanceTracker 性能追踪接口
type PerformanceTracker interface {
	// 操作追踪
	Track(operationName string) func()                              // 返回结束函数
	TrackWithContext(ctx context.Context, operationName string) func() // 带上下文追踪
	RecordDuration(operationName string, duration time.Duration)     // 直接记录耗时
	
	// 统计获取
	GetStats(operationName string) *PerformanceStats
	GetAllStats() map[string]*PerformanceStats
	
	// 管理操作
	Reset()
}

// MetricsCollector 指标收集器接口 - 异步批量指标处理
type MetricsCollector interface {
	// 生命周期
	Start()
	Stop()
	
	// 指标提交
	SubmitCounter(name string, value float64, labels map[string]string)
	SubmitGauge(name string, value float64, labels map[string]string)
	SubmitHistogram(name string, value float64, labels map[string]string)
	SubmitBatch(batch *MetricsBatch)
	
	// 状态查询
	BufferSize() int
	PendingCount() int
}

// MetricsMiddleware HTTP中间件接口
type MetricsMiddleware interface {
	WrapHandler(next http.Handler) http.Handler
	WrapHandlerFunc(next http.HandlerFunc) http.HandlerFunc
}

// PerformanceStats 性能统计信息
type PerformanceStats struct {
	Name      string        `json:"name"`
	Count     uint64        `json:"count"`
	TotalTime time.Duration `json:"total_time"`
	AvgTime   time.Duration `json:"avg_time"`
	MinTime   time.Duration `json:"min_time"`
	MaxTime   time.Duration `json:"max_time"`
	P50       time.Duration `json:"p50"`
	P95       time.Duration `json:"p95"`
	P99       time.Duration `json:"p99"`
}

// MetricsBatch 指标批次 - 用于批量提交
type MetricsBatch struct {
	Timestamp time.Time              `json:"timestamp"`
	Counters  map[string]float64     `json:"counters,omitempty"`
	Gauges    map[string]float64     `json:"gauges,omitempty"`
	Histogram map[string][]float64   `json:"histograms,omitempty"`
	Labels    map[string]string      `json:"labels,omitempty"`
}

// MetricsSnapshot 指标快照
type MetricsSnapshot struct {
	Timestamp time.Time                    `json:"timestamp"`
	Counters  map[string]float64           `json:"counters"`
	Gauges    map[string]float64           `json:"gauges"`
	Histogram map[string]*HistogramStats   `json:"histograms"`
}

// HistogramStats 直方图统计信息
type HistogramStats struct {
	Count   uint64  `json:"count"`
	Sum     float64 `json:"sum"`
	Average float64 `json:"average"`
	P50     float64 `json:"p50"`
	P95     float64 `json:"p95"`
	P99     float64 `json:"p99"`
}

// BucketInfo 桶信息
type BucketInfo struct {
	UpperBound float64 `json:"upper_bound"`
	Count      uint64  `json:"count"`
}

// HealthCheckMetrics 健康检查指标接口
type HealthCheckMetrics interface {
	RecordSuccess()
	RecordFailure()
	IsHealthy() bool
	GetStats() (success, failure float64, lastCheck time.Time, healthy bool)
	Reset()
}