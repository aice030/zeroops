package metrics

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusMetrics Prometheus指标实现
type PrometheusMetrics struct {
	config   Config
	registry *prometheus.Registry

	// 预创建的高频指标，避免运行时查找
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
	errorsTotal         *prometheus.CounterVec
	activeRequests      *prometheus.GaugeVec

	// 动态指标缓存
	mu         sync.RWMutex
	counters   map[string]*prometheus.CounterVec
	gauges     map[string]*prometheus.GaugeVec
	histograms map[string]*prometheus.HistogramVec

	// Timer池
	timerPool sync.Pool
}

// NewPrometheusMetrics 创建Prometheus指标实例
func NewPrometheusMetrics(config Config) *PrometheusMetrics {
	// 设置默认值
	if config.Namespace == "" {
		config.Namespace = "mock_s3"
	}
	if config.ServiceName == "" {
		config.ServiceName = "unknown"
	}
	if config.Labels == nil {
		config.Labels = make(map[string]string)
	}

	// 添加服务标签
	config.Labels["service"] = config.ServiceName
	config.Labels["version"] = config.ServiceVer

	registry := prometheus.NewRegistry()

	pm := &PrometheusMetrics{
		config:     config,
		registry:   registry,
		counters:   make(map[string]*prometheus.CounterVec),
		gauges:     make(map[string]*prometheus.GaugeVec),
		histograms: make(map[string]*prometheus.HistogramVec),
	}

	// 初始化Timer池
	pm.timerPool = sync.Pool{
		New: func() any {
			return &prometheusTimer{metrics: pm}
		},
	}

	pm.initPredefiedMetrics()
	return pm
}

// initPredefiedMetrics 初始化预定义的高频指标
func (pm *PrometheusMetrics) initPredefiedMetrics() {
	// HTTP请求总数
	pm.httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   pm.config.Namespace,
			Name:        "http_requests_total",
			Help:        "Total number of HTTP requests",
			ConstLabels: pm.config.Labels,
		},
		[]string{"method", "endpoint", "status"},
	)

	// HTTP请求耗时
	pm.httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace:   pm.config.Namespace,
			Name:        "http_request_duration_seconds",
			Help:        "HTTP request duration in seconds",
			ConstLabels: pm.config.Labels,
			Buckets:     []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "endpoint"},
	)

	// 错误总数
	pm.errorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   pm.config.Namespace,
			Name:        "errors_total",
			Help:        "Total number of errors",
			ConstLabels: pm.config.Labels,
		},
		[]string{"type", "operation"},
	)

	// 活跃请求数
	pm.activeRequests = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   pm.config.Namespace,
			Name:        "active_requests",
			Help:        "Number of active requests",
			ConstLabels: pm.config.Labels,
		},
		[]string{"endpoint"},
	)

	// 注册预定义指标
	pm.registry.MustRegister(
		pm.httpRequestsTotal,
		pm.httpRequestDuration,
		pm.errorsTotal,
		pm.activeRequests,
	)
}

// IncCounter 增加计数器
func (pm *PrometheusMetrics) IncCounter(name string, labels ...string) {
	counter := pm.getOrCreateCounter(name, labels...)
	counter.WithLabelValues(labels...).Inc()
}

// AddCounter 增加计数器指定值
func (pm *PrometheusMetrics) AddCounter(name string, value float64, labels ...string) {
	counter := pm.getOrCreateCounter(name, labels...)
	counter.WithLabelValues(labels...).Add(value)
}

// SetGauge 设置测量器值
func (pm *PrometheusMetrics) SetGauge(name string, value float64, labels ...string) {
	gauge := pm.getOrCreateGauge(name, labels...)
	gauge.WithLabelValues(labels...).Set(value)
}

// IncGauge 增加测量器
func (pm *PrometheusMetrics) IncGauge(name string, labels ...string) {
	gauge := pm.getOrCreateGauge(name, labels...)
	gauge.WithLabelValues(labels...).Inc()
}

// DecGauge 减少测量器
func (pm *PrometheusMetrics) DecGauge(name string, labels ...string) {
	gauge := pm.getOrCreateGauge(name, labels...)
	gauge.WithLabelValues(labels...).Dec()
}

// ObserveHistogram 记录直方图观测值
func (pm *PrometheusMetrics) ObserveHistogram(name string, value float64, labels ...string) {
	histogram := pm.getOrCreateHistogram(name, labels...)
	histogram.WithLabelValues(labels...).Observe(value)
}

// ObserveDuration 记录耗时
func (pm *PrometheusMetrics) ObserveDuration(name string, duration time.Duration, labels ...string) {
	pm.ObserveHistogram(name, duration.Seconds(), labels...)
}

// StartTimer 开始计时
func (pm *PrometheusMetrics) StartTimer(name string, labels ...string) Timer {
	timer := pm.timerPool.Get().(*prometheusTimer)
	timer.name = name
	timer.labels = labels
	timer.startTime = time.Now()
	return timer
}

// RecordHTTPRequest 记录HTTP请求（高频操作优化）
func (pm *PrometheusMetrics) RecordHTTPRequest(method, endpoint, status string, duration time.Duration) {
	// 直接使用预创建的指标，避免查找开销
	pm.httpRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
	pm.httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// RecordError 记录错误
func (pm *PrometheusMetrics) RecordError(errorType, operation string) {
	pm.errorsTotal.WithLabelValues(errorType, operation).Inc()
}

// HTTPHandler 返回Prometheus处理器
func (pm *PrometheusMetrics) HTTPHandler() http.Handler {
	return promhttp.HandlerFor(pm.registry, promhttp.HandlerOpts{
		ErrorHandling: promhttp.ContinueOnError,
		Registry:      pm.registry,
	})
}

// Middleware 返回HTTP中间件
func (pm *PrometheusMetrics) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			endpoint := r.URL.Path
			method := r.Method

			// 增加活跃请求数
			pm.activeRequests.WithLabelValues(endpoint).Inc()
			defer pm.activeRequests.WithLabelValues(endpoint).Dec()

			// 包装ResponseWriter捕获状态码
			wrapped := &metricsResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// 执行处理器
			next.ServeHTTP(wrapped, r)

			// 记录指标
			duration := time.Since(start)
			status := fmt.Sprintf("%d", wrapped.statusCode)
			pm.RecordHTTPRequest(method, endpoint, status, duration)
		})
	}
}

// Shutdown 优雅关闭
func (pm *PrometheusMetrics) Shutdown(ctx context.Context) error {
	// Prometheus不需要特殊清理
	return nil
}

// getOrCreateCounter 获取或创建计数器
func (pm *PrometheusMetrics) getOrCreateCounter(name string, labelNames ...string) *prometheus.CounterVec {
	pm.mu.RLock()
	counter, exists := pm.counters[name]
	pm.mu.RUnlock()

	if exists {
		return counter
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// 双检查锁定
	if counter, exists := pm.counters[name]; exists {
		return counter
	}

	counter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   pm.config.Namespace,
			Name:        name,
			Help:        fmt.Sprintf("Counter metric: %s", name),
			ConstLabels: pm.config.Labels,
		},
		labelNames,
	)

	pm.registry.MustRegister(counter)
	pm.counters[name] = counter
	return counter
}

// getOrCreateGauge 获取或创建测量器
func (pm *PrometheusMetrics) getOrCreateGauge(name string, labelNames ...string) *prometheus.GaugeVec {
	pm.mu.RLock()
	gauge, exists := pm.gauges[name]
	pm.mu.RUnlock()

	if exists {
		return gauge
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	if gauge, exists := pm.gauges[name]; exists {
		return gauge
	}

	gauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   pm.config.Namespace,
			Name:        name,
			Help:        fmt.Sprintf("Gauge metric: %s", name),
			ConstLabels: pm.config.Labels,
		},
		labelNames,
	)

	pm.registry.MustRegister(gauge)
	pm.gauges[name] = gauge
	return gauge
}

// getOrCreateHistogram 获取或创建直方图
func (pm *PrometheusMetrics) getOrCreateHistogram(name string, labelNames ...string) *prometheus.HistogramVec {
	pm.mu.RLock()
	histogram, exists := pm.histograms[name]
	pm.mu.RUnlock()

	if exists {
		return histogram
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	if histogram, exists := pm.histograms[name]; exists {
		return histogram
	}

	histogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace:   pm.config.Namespace,
			Name:        name,
			Help:        fmt.Sprintf("Histogram metric: %s", name),
			ConstLabels: pm.config.Labels,
			Buckets:     prometheus.DefBuckets,
		},
		labelNames,
	)

	pm.registry.MustRegister(histogram)
	pm.histograms[name] = histogram
	return histogram
}

// prometheusTimer Timer实现
type prometheusTimer struct {
	metrics   *PrometheusMetrics
	name      string
	labels    []string
	startTime time.Time
}

// ObserveDuration 记录耗时并返回到对象池
func (t *prometheusTimer) ObserveDuration() time.Duration {
	duration := time.Since(t.startTime)
	t.metrics.ObserveDuration(t.name, duration, t.labels...)

	// 清理并返回对象池
	t.name = ""
	t.labels = nil
	t.metrics.timerPool.Put(t)

	return duration
}

// metricsResponseWriter 包装器用于捕获HTTP状态码
type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *metricsResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// New 创建默认的Prometheus指标实例
func New(serviceName, version string) Metrics {
	config := Config{
		ServiceName: serviceName,
		ServiceVer:  version,
		Namespace:   "mock_s3",
		Labels:      make(map[string]string),
	}
	return NewPrometheusMetrics(config)
}

// NewWithConfig 使用配置创建指标实例
func NewWithConfig(config Config) Metrics {
	return NewPrometheusMetrics(config)
}
