package metrics

import (
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// 基于原子操作的高性能指标实现

// atomicCounter 基于原子操作的高性能计数器
type atomicCounter struct {
	value uint64 // 使用uint64存储，通过位操作转换float64
	labels map[string]string
}

// NewAtomicCounter 创建原子计数器
func NewAtomicCounter() Counter {
	return &atomicCounter{
		labels: make(map[string]string),
	}
}

// Inc 原子增加1
func (c *atomicCounter) Inc() {
	c.Add(1)
}

// Add 原子增加指定值
func (c *atomicCounter) Add(value float64) {
	if value < 0 {
		return // 计数器不能减少
	}
	
	// 将float64转换为uint64进行原子操作
	bits := atomic.AddUint64(&c.value, floatToBits(value))
	_ = bitsToFloat(bits) // 转换验证
}

// Get 获取当前值
func (c *atomicCounter) Get() float64 {
	bits := atomic.LoadUint64(&c.value)
	return bitsToFloat(bits)
}

// Reset 重置计数器
func (c *atomicCounter) Reset() {
	atomic.StoreUint64(&c.value, 0)
}

// WithLabels 创建带标签的计数器副本
func (c *atomicCounter) WithLabels(labels map[string]string) Counter {
	newLabels := make(map[string]string, len(c.labels)+len(labels))
	for k, v := range c.labels {
		newLabels[k] = v
	}
	for k, v := range labels {
		newLabels[k] = v
	}
	
	return &atomicCounter{
		value:  atomic.LoadUint64(&c.value),
		labels: newLabels,
	}
}

// atomicGauge 基于原子操作的高性能测量器
type atomicGauge struct {
	value uint64 // 使用uint64存储float64
	labels map[string]string
}

// NewAtomicGauge 创建原子测量器
func NewAtomicGauge() Gauge {
	return &atomicGauge{
		labels: make(map[string]string),
	}
}

// Inc 原子增加1
func (g *atomicGauge) Inc() {
	g.Add(1)
}

// Dec 原子减少1
func (g *atomicGauge) Dec() {
	g.Sub(1)
}

// Add 原子增加指定值
func (g *atomicGauge) Add(value float64) {
	for {
		old := atomic.LoadUint64(&g.value)
		oldVal := bitsToFloat(old)
		newVal := oldVal + value
		new := floatToBits(newVal)
		
		if atomic.CompareAndSwapUint64(&g.value, old, new) {
			break
		}
	}
}

// Sub 原子减少指定值
func (g *atomicGauge) Sub(value float64) {
	g.Add(-value)
}

// Set 原子设置值
func (g *atomicGauge) Set(value float64) {
	atomic.StoreUint64(&g.value, floatToBits(value))
}

// Get 获取当前值
func (g *atomicGauge) Get() float64 {
	bits := atomic.LoadUint64(&g.value)
	return bitsToFloat(bits)
}

// WithLabels 创建带标签的测量器副本
func (g *atomicGauge) WithLabels(labels map[string]string) Gauge {
	newLabels := make(map[string]string, len(g.labels)+len(labels))
	for k, v := range g.labels {
		newLabels[k] = v
	}
	for k, v := range labels {
		newLabels[k] = v
	}
	
	return &atomicGauge{
		value:  atomic.LoadUint64(&g.value),
		labels: newLabels,
	}
}

// 工具函数：float64和uint64之间的转换
func floatToBits(f float64) uint64 {
	return *(*uint64)(unsafe.Pointer(&f))
}

func bitsToFloat(b uint64) float64 {
	return *(*float64)(unsafe.Pointer(&b))
}


// LabeledCounter 带标签的计数器管理器
type LabeledCounter struct {
	mu       sync.RWMutex
	counters map[string]Counter
	factory  func() Counter
}

// NewLabeledCounter 创建标签计数器管理器
func NewLabeledCounter() *LabeledCounter {
	return &LabeledCounter{
		counters: make(map[string]Counter),
		factory:  NewAtomicCounter,
	}
}

// WithLabelValues 根据标签值获取计数器
func (lc *LabeledCounter) WithLabelValues(labelValues ...string) Counter {
	key := generateKey(labelValues...)
	
	lc.mu.RLock()
	counter, exists := lc.counters[key]
	lc.mu.RUnlock()
	
	if exists {
		return counter
	}
	
	lc.mu.Lock()
	// 双检查锁定模式
	if counter, exists := lc.counters[key]; exists {
		lc.mu.Unlock()
		return counter
	}
	
	counter = lc.factory()
	lc.counters[key] = counter
	lc.mu.Unlock()
	
	return counter
}

// Reset 重置所有计数器
func (lc *LabeledCounter) Reset() {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	
	for _, counter := range lc.counters {
		counter.Reset()
	}
}

// GetAll 获取所有计数器及其值
func (lc *LabeledCounter) GetAll() map[string]float64 {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	
	result := make(map[string]float64, len(lc.counters))
	for key, counter := range lc.counters {
		result[key] = counter.Get()
	}
	
	return result
}

// LabeledGauge 带标签的测量器管理器
type LabeledGauge struct {
	mu     sync.RWMutex
	gauges map[string]Gauge
	factory func() Gauge
}

// NewLabeledGauge 创建标签测量器管理器
func NewLabeledGauge() *LabeledGauge {
	return &LabeledGauge{
		gauges:  make(map[string]Gauge),
		factory: NewAtomicGauge,
	}
}

// WithLabelValues 根据标签值获取测量器
func (lg *LabeledGauge) WithLabelValues(labelValues ...string) Gauge {
	key := generateKey(labelValues...)
	
	lg.mu.RLock()
	gauge, exists := lg.gauges[key]
	lg.mu.RUnlock()
	
	if exists {
		return gauge
	}
	
	lg.mu.Lock()
	// 双检查锁定模式
	if gauge, exists := lg.gauges[key]; exists {
		lg.mu.Unlock()
		return gauge
	}
	
	gauge = lg.factory()
	lg.gauges[key] = gauge
	lg.mu.Unlock()
	
	return gauge
}

// GetAll 获取所有测量器及其值
func (lg *LabeledGauge) GetAll() map[string]float64 {
	lg.mu.RLock()
	defer lg.mu.RUnlock()
	
	result := make(map[string]float64, len(lg.gauges))
	for key, gauge := range lg.gauges {
		result[key] = gauge.Get()
	}
	
	return result
}

// RateCounter 速率计数器，计算每秒速率
type RateCounter struct {
	counter   Counter
	lastValue float64
	lastTime  time.Time
	mu        sync.RWMutex
}

// NewRateCounter 创建速率计数器
func NewRateCounter() *RateCounter {
	return &RateCounter{
		counter:  NewAtomicCounter(),
		lastTime: time.Now(),
	}
}

// Inc 增加计数
func (rc *RateCounter) Inc() {
	rc.counter.Inc()
}

// Add 增加指定值
func (rc *RateCounter) Add(value float64) {
	rc.counter.Add(value)
}

// Rate 获取当前速率（每秒）
func (rc *RateCounter) Rate() float64 {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	
	now := time.Now()
	currentValue := rc.counter.Get()
	
	if rc.lastTime.IsZero() {
		rc.lastTime = now
		rc.lastValue = currentValue
		return 0
	}
	
	duration := now.Sub(rc.lastTime).Seconds()
	if duration <= 0 {
		return 0
	}
	
	rate := (currentValue - rc.lastValue) / duration
	rc.lastValue = currentValue
	rc.lastTime = now
	
	return rate
}

// Get 获取总计数值
func (rc *RateCounter) Get() float64 {
	return rc.counter.Get()
}

// MovingAverage 移动平均计算器
type MovingAverage struct {
	values []float64
	size   int
	index  int
	count  int
	sum    float64
	mu     sync.RWMutex
}

// NewMovingAverage 创建移动平均计算器
func NewMovingAverage(size int) *MovingAverage {
	if size <= 0 {
		size = 10
	}
	return &MovingAverage{
		values: make([]float64, size),
		size:   size,
	}
}

// Add 添加新值
func (ma *MovingAverage) Add(value float64) {
	ma.mu.Lock()
	defer ma.mu.Unlock()
	
	if ma.count < ma.size {
		ma.sum += value
		ma.values[ma.index] = value
		ma.count++
	} else {
		// 移除最旧的值，添加新值
		ma.sum = ma.sum - ma.values[ma.index] + value
		ma.values[ma.index] = value
	}
	
	ma.index = (ma.index + 1) % ma.size
}

// Average 获取移动平均值
func (ma *MovingAverage) Average() float64 {
	ma.mu.RLock()
	defer ma.mu.RUnlock()
	
	if ma.count == 0 {
		return 0
	}
	
	return ma.sum / float64(ma.count)
}

// Count 获取样本数量
func (ma *MovingAverage) Count() int {
	ma.mu.RLock()
	defer ma.mu.RUnlock()
	return ma.count
}

// Reset 重置移动平均计算器
func (ma *MovingAverage) Reset() {
	ma.mu.Lock()
	defer ma.mu.Unlock()
	
	ma.index = 0
	ma.count = 0
	ma.sum = 0
	for i := range ma.values {
		ma.values[i] = 0
	}
}

// 工具函数

// generateKey 生成标签键
func generateKey(values ...string) string {
	if len(values) == 0 {
		return ""
	}
	
	if len(values) == 1 {
		return values[0]
	}
	
	// 预计算总长度，避免多次内存分配
	totalLen := len(values) - 1 // 分隔符数量
	for _, v := range values {
		totalLen += len(v)
	}
	
	// 使用strings.Builder高效拼接
	var result strings.Builder
	result.Grow(totalLen)
	
	result.WriteString(values[0])
	for _, v := range values[1:] {
		result.WriteByte('|')
		result.WriteString(v)
	}
	
	return result.String()
}

// atomicMetricsSnapshot 原子指标快照
type atomicMetricsSnapshot struct {
	Timestamp time.Time              `json:"timestamp"`
	Counters  map[string]float64     `json:"counters"`
	Gauges    map[string]float64     `json:"gauges"`
}

// TakeSnapshot 获取指标快照
func TakeAtomicSnapshot(counters *LabeledCounter, gauges *LabeledGauge) *atomicMetricsSnapshot {
	snapshot := &atomicMetricsSnapshot{
		Timestamp: time.Now(),
		Counters:  make(map[string]float64),
		Gauges:    make(map[string]float64),
	}
	
	if counters != nil {
		snapshot.Counters = counters.GetAll()
	}
	
	if gauges != nil {
		snapshot.Gauges = gauges.GetAll()
	}
	
	return snapshot
}

// HealthCheck 健康检查指标
type HealthCheck struct {
	successCounter Counter
	failureCounter Counter
	lastCheckTime  int64 // 使用原子操作
	status         int32 // 0: unknown, 1: healthy, 2: unhealthy
}

// NewHealthCheck 创建健康检查指标
func NewHealthCheck() *HealthCheck {
	return &HealthCheck{
		successCounter: NewAtomicCounter(),
		failureCounter: NewAtomicCounter(),
	}
}

// RecordSuccess 记录成功检查
func (hc *HealthCheck) RecordSuccess() {
	hc.successCounter.Inc()
	atomic.StoreInt64(&hc.lastCheckTime, time.Now().Unix())
	atomic.StoreInt32(&hc.status, 1) // healthy
}

// RecordFailure 记录失败检查
func (hc *HealthCheck) RecordFailure() {
	hc.failureCounter.Inc()
	atomic.StoreInt64(&hc.lastCheckTime, time.Now().Unix())
	atomic.StoreInt32(&hc.status, 2) // unhealthy
}

// IsHealthy 检查是否健康
func (hc *HealthCheck) IsHealthy() bool {
	return atomic.LoadInt32(&hc.status) == 1
}

// GetStats 获取健康检查统计
func (hc *HealthCheck) GetStats() (success, failure float64, lastCheck time.Time, healthy bool) {
	success = hc.successCounter.Get()
	failure = hc.failureCounter.Get()
	lastCheck = time.Unix(atomic.LoadInt64(&hc.lastCheckTime), 0)
	healthy = hc.IsHealthy()
	return
}