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

// Prometheus指标实现

// PrometheusConfig Prometheus配置
type PrometheusConfig struct {
	ServiceName string            // 服务名称
	ServiceVer  string            // 服务版本
	Namespace   string            // 指标命名空间
	Subsystem   string            // 子系统名称
	Labels      map[string]string // 默认标签
	Registry    *prometheus.Registry // 自定义注册表，nil时使用默认
}

// PrometheusProvider Prometheus指标提供者实现
type PrometheusProvider struct {
	config   PrometheusConfig
	registry *prometheus.Registry
	
	// 预定义基础指标
	requestTotal    *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	errorTotal      *prometheus.CounterVec
	activeRequests  *prometheus.GaugeVec
	
	// 动态业务指标
	mu                 sync.RWMutex
	businessCounters   map[string]*prometheus.CounterVec
	businessGauges     map[string]*prometheus.GaugeVec
	businessHistograms map[string]*prometheus.HistogramVec
	businessSummaries  map[string]*prometheus.SummaryVec
}

// NewPrometheusProvider 创建Prometheus指标管理器
func NewPrometheusProvider(config PrometheusConfig) *PrometheusProvider {
	// 设置默认值
	if config.Namespace == "" {
		config.Namespace = "mock_s3"
	}
	if config.ServiceName == "" {
		config.ServiceName = "unknown"
	}
	if config.ServiceVer == "" {
		config.ServiceVer = "unknown"
	}
	if config.Labels == nil {
		config.Labels = make(map[string]string)
	}
	
	// 添加默认标签
	config.Labels["service"] = config.ServiceName
	config.Labels["version"] = config.ServiceVer
	
	// 使用自定义或默认注册表
	registry := config.Registry
	if registry == nil {
		registry = prometheus.NewRegistry()
	}
	
	pm := &PrometheusProvider{
		config:             config,
		registry:           registry,
		businessCounters:   make(map[string]*prometheus.CounterVec),
		businessGauges:     make(map[string]*prometheus.GaugeVec),
		businessHistograms: make(map[string]*prometheus.HistogramVec),
		businessSummaries:  make(map[string]*prometheus.SummaryVec),
	}
	
	pm.initBasicMetrics()
	return pm
}

// initBasicMetrics 初始化基础指标
func (pm *PrometheusProvider) initBasicMetrics() {
	// HTTP请求总数
	pm.requestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: pm.config.Namespace,
			Subsystem: pm.config.Subsystem,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
			ConstLabels: pm.config.Labels,
		},
		[]string{"method", "endpoint", "status_code"},
	)
	
	// HTTP请求耗时分布
	pm.requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: pm.config.Namespace,
			Subsystem: pm.config.Subsystem,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			ConstLabels: pm.config.Labels,
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "endpoint"},
	)
	
	// 错误总数
	pm.errorTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: pm.config.Namespace,
			Subsystem: pm.config.Subsystem,
			Name:      "errors_total",
			Help:      "Total number of errors",
			ConstLabels: pm.config.Labels,
		},
		[]string{"error_type", "error_code", "operation"},
	)
	
	// 活跃请求数
	pm.activeRequests = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: pm.config.Namespace,
			Subsystem: pm.config.Subsystem,
			Name:      "active_requests",
			Help:      "Number of active requests",
			ConstLabels: pm.config.Labels,
		},
		[]string{"endpoint"},
	)
	
	// 注册基础指标
	pm.registry.MustRegister(
		pm.requestTotal,
		pm.requestDuration,
		pm.errorTotal,
		pm.activeRequests,
	)
}

// GetRegistry 获取Prometheus注册表
func (pm *PrometheusProvider) GetRegistry() *prometheus.Registry {
	return pm.registry
}

// HTTPHandler 返回Prometheus指标的HTTP处理器
func (pm *PrometheusProvider) HTTPHandler() http.Handler {
	return promhttp.HandlerFor(pm.registry, promhttp.HandlerOpts{
		ErrorLog:      nil, // 使用默认错误日志
		ErrorHandling: promhttp.ContinueOnError,
		Registry:      pm.registry,
	})
}

// StartHTTPServer 启动Prometheus指标HTTP服务器
func (pm *PrometheusProvider) StartHTTPServer(addr string) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", pm.HTTPHandler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	fmt.Printf("Prometheus metrics server starting on %s\n", addr)
	return http.ListenAndServe(addr, mux)
}

// 基础指标操作方法

// IncRequestTotal 增加请求总数
func (pm *PrometheusProvider) IncRequestTotal(method, endpoint, statusCode string) {
	pm.requestTotal.WithLabelValues(method, endpoint, statusCode).Inc()
}

// ObserveRequestDuration 记录请求耗时
func (pm *PrometheusProvider) ObserveRequestDuration(method, endpoint string, duration time.Duration) {
	pm.requestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// IncErrorTotal 增加错误总数
func (pm *PrometheusProvider) IncErrorTotal(errorType, errorCode, operation string) {
	pm.errorTotal.WithLabelValues(errorType, errorCode, operation).Inc()
}

// IncActiveRequests 增加活跃请求数
func (pm *PrometheusProvider) IncActiveRequests(endpoint string) {
	pm.activeRequests.WithLabelValues(endpoint).Inc()
}

// DecActiveRequests 减少活跃请求数
func (pm *PrometheusProvider) DecActiveRequests(endpoint string) {
	pm.activeRequests.WithLabelValues(endpoint).Dec()
}

// 动态指标创建和操作

// NewCounter 创建新的计数器指标
func (pm *PrometheusProvider) NewCounter(name, help string, labelNames []string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	if _, exists := pm.businessCounters[name]; exists {
		return fmt.Errorf("counter %s already exists", name)
	}
	
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: pm.config.Namespace,
			Subsystem: pm.config.Subsystem,
			Name:      name,
			Help:      help,
			ConstLabels: pm.config.Labels,
		},
		labelNames,
	)
	
	pm.registry.MustRegister(counter)
	pm.businessCounters[name] = counter
	return nil
}

// IncCounter 增加计数器指标
func (pm *PrometheusProvider) IncCounter(name string, labelValues ...string) error {
	pm.mu.RLock()
	counter, exists := pm.businessCounters[name]
	pm.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("counter %s not found", name)
	}
	
	counter.WithLabelValues(labelValues...).Inc()
	return nil
}

// AddCounter 增加计数器指标指定值
func (pm *PrometheusProvider) AddCounter(name string, value float64, labelValues ...string) error {
	counter, exists := pm.businessCounters[name]
	if !exists {
		return fmt.Errorf("counter %s not found", name)
	}
	
	counter.WithLabelValues(labelValues...).Add(value)
	return nil
}

// NewGauge 创建新的测量器指标
func (pm *PrometheusProvider) NewGauge(name, help string, labelNames []string) error {
	if _, exists := pm.businessGauges[name]; exists {
		return fmt.Errorf("gauge %s already exists", name)
	}
	
	gauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: pm.config.Namespace,
			Subsystem: pm.config.Subsystem,
			Name:      name,
			Help:      help,
			ConstLabels: pm.config.Labels,
		},
		labelNames,
	)
	
	pm.registry.MustRegister(gauge)
	pm.businessGauges[name] = gauge
	return nil
}

// SetGauge 设置测量器指标值
func (pm *PrometheusProvider) SetGauge(name string, value float64, labelValues ...string) error {
	gauge, exists := pm.businessGauges[name]
	if !exists {
		return fmt.Errorf("gauge %s not found", name)
	}
	
	gauge.WithLabelValues(labelValues...).Set(value)
	return nil
}

// IncGauge 增加测量器指标值
func (pm *PrometheusProvider) IncGauge(name string, labelValues ...string) error {
	gauge, exists := pm.businessGauges[name]
	if !exists {
		return fmt.Errorf("gauge %s not found", name)
	}
	
	gauge.WithLabelValues(labelValues...).Inc()
	return nil
}

// DecGauge 减少测量器指标值
func (pm *PrometheusProvider) DecGauge(name string, labelValues ...string) error {
	gauge, exists := pm.businessGauges[name]
	if !exists {
		return fmt.Errorf("gauge %s not found", name)
	}
	
	gauge.WithLabelValues(labelValues...).Dec()
	return nil
}

// NewHistogram 创建新的直方图指标
func (pm *PrometheusProvider) NewHistogram(name, help string, buckets []float64, labelNames []string) error {
	if _, exists := pm.businessHistograms[name]; exists {
		return fmt.Errorf("histogram %s already exists", name)
	}
	
	histogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: pm.config.Namespace,
			Subsystem: pm.config.Subsystem,
			Name:      name,
			Help:      help,
			ConstLabels: pm.config.Labels,
			Buckets: buckets,
		},
		labelNames,
	)
	
	pm.registry.MustRegister(histogram)
	pm.businessHistograms[name] = histogram
	return nil
}

// ObserveHistogram 记录直方图观测值
func (pm *PrometheusProvider) ObserveHistogram(name string, value float64, labelValues ...string) error {
	histogram, exists := pm.businessHistograms[name]
	if !exists {
		return fmt.Errorf("histogram %s not found", name)
	}
	
	histogram.WithLabelValues(labelValues...).Observe(value)
	return nil
}

// PrometheusTimer 计时器结构，用于自动记录耗时到Prometheus
type PrometheusTimer struct {
	metrics   *PrometheusProvider
	name      string
	labels    []string
	startTime time.Time
}

// NewTimer 创建新的计时器
func (pm *PrometheusProvider) NewTimer(name string, labelValues ...string) *PrometheusTimer {
	return &PrometheusTimer{
		metrics:   pm,
		name:      name,
		labels:    labelValues,
		startTime: time.Now(),
	}
}

// ObserveDuration 记录耗时到直方图
func (t *PrometheusTimer) ObserveDuration() time.Duration {
	duration := time.Since(t.startTime)
	if err := t.metrics.ObserveHistogram(t.name, duration.Seconds(), t.labels...); err != nil {
		// 忽略错误，避免影响业务逻辑
	}
	return duration
}

// Middleware 中间件接口
type Middleware interface {
	WrapHandler(next http.Handler) http.Handler
}

// HTTPMiddleware HTTP指标中间件
type HTTPMiddleware struct {
	metrics *PrometheusProvider
}

// NewHTTPMiddleware 创建HTTP指标中间件
func (pm *PrometheusProvider) NewHTTPMiddleware() *HTTPMiddleware {
	return &HTTPMiddleware{metrics: pm}
}

// WrapHandler 包装HTTP处理器，自动记录指标
func (m *HTTPMiddleware) WrapHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		endpoint := r.URL.Path
		method := r.Method
		
		// 增加活跃请求数
		m.metrics.IncActiveRequests(endpoint)
		defer m.metrics.DecActiveRequests(endpoint)
		
		// 包装ResponseWriter以捕获状态码
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		// 执行处理器
		next.ServeHTTP(wrapped, r)
		
		// 记录指标
		duration := time.Since(start)
		statusCode := fmt.Sprintf("%d", wrapped.statusCode)
		
		m.metrics.IncRequestTotal(method, endpoint, statusCode)
		m.metrics.ObserveRequestDuration(method, endpoint, duration)
	})
}

// responseWriter 包装器，用于捕获HTTP状态码
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// 健康检查和指标重置

// Reset 重置所有指标（谨慎使用）
func (pm *PrometheusProvider) Reset() {
	// 重置基础指标
	pm.requestTotal.Reset()
	pm.requestDuration.Reset()
	pm.errorTotal.Reset()
	
	// 重置业务指标
	for _, counter := range pm.businessCounters {
		counter.Reset()
	}
	for _, histogram := range pm.businessHistograms {
		histogram.Reset()
	}
	for _, summary := range pm.businessSummaries {
		summary.Reset()
	}
}

// GetMetricNames 获取所有指标名称
func (pm *PrometheusProvider) GetMetricNames() []string {
	names := []string{
		"http_requests_total",
		"http_request_duration_seconds",
		"errors_total", 
		"active_requests",
	}
	
	for name := range pm.businessCounters {
		names = append(names, name)
	}
	for name := range pm.businessGauges {
		names = append(names, name)
	}
	for name := range pm.businessHistograms {
		names = append(names, name)
	}
	for name := range pm.businessSummaries {
		names = append(names, name)
	}
	
	return names
}

// Shutdown 优雅关闭，清理资源
func (pm *PrometheusProvider) Shutdown(ctx context.Context) error {
	// Prometheus客户端库通常不需要显式清理
	// 这里可以添加自定义的清理逻辑
	return nil
}