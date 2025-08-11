package metrics

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// FastMetrics 高性能指标管理器，避免动态查找开销
type FastMetrics struct {
	// 预创建的高频指标，避免WithLabelValues查找
	httpRequestsTotal   prometheus.Counter
	httpRequest200Total prometheus.Counter
	httpRequest404Total prometheus.Counter
	httpRequest500Total prometheus.Counter

	// 预创建的耗时指标
	httpGetDuration    prometheus.Observer
	httpPostDuration   prometheus.Observer
	httpPutDuration    prometheus.Observer
	httpDeleteDuration prometheus.Observer

	// 原子计数器用于高频简单指标
	totalRequests     uint64
	totalErrors       uint64
	activeConnections int64

	// 性能计时器池
	timerPool sync.Pool
}

// NewFastMetrics 创建高性能指标管理器
func NewFastMetrics(registry *prometheus.Registry, serviceName string) *FastMetrics {
	// 预创建常用指标的特定标签值实例
	requestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "app",
			Name:      "http_requests_total",
			Help:      "Total HTTP requests",
		},
		[]string{"method", "status"},
	)

	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "app",
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		},
		[]string{"method"},
	)

	registry.MustRegister(requestsTotal, requestDuration)

	fm := &FastMetrics{
		// 预创建常用的标签组合，避免运行时查找
		httpRequestsTotal:   requestsTotal.WithLabelValues("", ""),
		httpRequest200Total: requestsTotal.WithLabelValues("GET", "200"),
		httpRequest404Total: requestsTotal.WithLabelValues("GET", "404"),
		httpRequest500Total: requestsTotal.WithLabelValues("POST", "500"),

		httpGetDuration:    requestDuration.WithLabelValues("GET"),
		httpPostDuration:   requestDuration.WithLabelValues("POST"),
		httpPutDuration:    requestDuration.WithLabelValues("PUT"),
		httpDeleteDuration: requestDuration.WithLabelValues("DELETE"),
	}

	// 初始化Timer对象池
	fm.timerPool = sync.Pool{
		New: func() any {
			return &FastTimer{
				startTime: time.Now(),
			}
		},
	}

	return fm
}

// FastTimer 高性能计时器
type FastTimer struct {
	startTime time.Time
	observer  prometheus.Observer
}

// StartTimer 开始计时（从对象池获取）
func (fm *FastMetrics) StartTimer(observer prometheus.Observer) *FastTimer {
	timer := fm.timerPool.Get().(*FastTimer)
	timer.startTime = time.Now()
	timer.observer = observer
	return timer
}

// ObserveDuration 记录耗时并返回到对象池
func (t *FastTimer) ObserveDuration(fm *FastMetrics) time.Duration {
	duration := time.Since(t.startTime)
	if t.observer != nil {
		t.observer.Observe(duration.Seconds())
	}

	// 返回到对象池
	t.observer = nil
	fm.timerPool.Put(t)

	return duration
}

// 高频操作的快速方法 - 使用原子操作

// IncTotalRequests 增加总请求数（原子操作）
func (fm *FastMetrics) IncTotalRequests() {
	atomic.AddUint64(&fm.totalRequests, 1)
}

// IncTotalErrors 增加总错误数（原子操作）
func (fm *FastMetrics) IncTotalErrors() {
	atomic.AddUint64(&fm.totalErrors, 1)
}

// AddActiveConnections 增加活跃连接数
func (fm *FastMetrics) AddActiveConnections(delta int64) {
	atomic.AddInt64(&fm.activeConnections, delta)
}

// GetTotalRequests 获取总请求数
func (fm *FastMetrics) GetTotalRequests() uint64 {
	return atomic.LoadUint64(&fm.totalRequests)
}

// GetTotalErrors 获取总错误数
func (fm *FastMetrics) GetTotalErrors() uint64 {
	return atomic.LoadUint64(&fm.totalErrors)
}

// GetActiveConnections 获取活跃连接数
func (fm *FastMetrics) GetActiveConnections() int64 {
	return atomic.LoadInt64(&fm.activeConnections)
}

// 预创建指标的快速访问方法

// IncHTTP200 增加200响应计数
func (fm *FastMetrics) IncHTTP200() {
	fm.httpRequest200Total.Inc()
}

// IncHTTP404 增加404响应计数
func (fm *FastMetrics) IncHTTP404() {
	fm.httpRequest404Total.Inc()
}

// IncHTTP500 增加500响应计数
func (fm *FastMetrics) IncHTTP500() {
	fm.httpRequest500Total.Inc()
}

// ObserveGETDuration 记录GET请求耗时
func (fm *FastMetrics) ObserveGETDuration(duration time.Duration) {
	fm.httpGetDuration.Observe(duration.Seconds())
}

// ObservePOSTDuration 记录POST请求耗时
func (fm *FastMetrics) ObservePOSTDuration(duration time.Duration) {
	fm.httpPostDuration.Observe(duration.Seconds())
}

// BatchMetrics 批量指标更新，减少函数调用开销
type BatchMetrics struct {
	HTTP200Count int
	HTTP404Count int
	HTTP500Count int
	ErrorCount   uint64
	RequestCount uint64
}

// BatchUpdate 批量更新指标
func (fm *FastMetrics) BatchUpdate(batch BatchMetrics) {
	// 批量更新Prometheus指标
	for i := 0; i < batch.HTTP200Count; i++ {
		fm.httpRequest200Total.Inc()
	}
	for i := 0; i < batch.HTTP404Count; i++ {
		fm.httpRequest404Total.Inc()
	}
	for i := 0; i < batch.HTTP500Count; i++ {
		fm.httpRequest500Total.Inc()
	}

	// 批量更新原子计数器
	if batch.RequestCount > 0 {
		atomic.AddUint64(&fm.totalRequests, batch.RequestCount)
	}
	if batch.ErrorCount > 0 {
		atomic.AddUint64(&fm.totalErrors, batch.ErrorCount)
	}
}

// FastMetricsCollector 高性能指标收集器
type FastMetricsCollector struct {
	fastMetrics *FastMetrics
	buffer      chan BatchMetrics
	done        chan struct{}
	bufferSize  int
}

// NewFastMetricsCollector 创建指标收集器
func NewFastMetricsCollector(fastMetrics *FastMetrics, bufferSize int) *FastMetricsCollector {
	return &FastMetricsCollector{
		fastMetrics: fastMetrics,
		buffer:      make(chan BatchMetrics, bufferSize),
		done:        make(chan struct{}),
		bufferSize:  bufferSize,
	}
}

// Start 启动异步指标收集
func (mc *FastMetricsCollector) Start() {
	go mc.collectLoop()
}

// Stop 停止指标收集
func (mc *FastMetricsCollector) Stop() {
	close(mc.done)
}

// Submit 提交指标批次
func (mc *FastMetricsCollector) Submit(batch BatchMetrics) {
	select {
	case mc.buffer <- batch:
		// 成功提交
	default:
		// 缓冲区满，直接更新避免阻塞
		mc.fastMetrics.BatchUpdate(batch)
	}
}

// collectLoop 收集循环
func (mc *FastMetricsCollector) collectLoop() {
	ticker := time.NewTicker(100 * time.Millisecond) // 100ms批次间隔
	defer ticker.Stop()

	for {
		select {
		case batch := <-mc.buffer:
			mc.fastMetrics.BatchUpdate(batch)

		case <-ticker.C:
			// 定期处理可能积压的批次
			mc.drainBuffer()

		case <-mc.done:
			// 处理剩余批次
			mc.drainBuffer()
			return
		}
	}
}

// drainBuffer 排空缓冲区
func (mc *FastMetricsCollector) drainBuffer() {
	for {
		select {
		case batch := <-mc.buffer:
			mc.fastMetrics.BatchUpdate(batch)
		default:
			return
		}
	}
}
