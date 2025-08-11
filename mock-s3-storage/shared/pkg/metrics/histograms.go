package metrics

import (
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// 直方图具体实现

// BucketHistogram 基于桶的直方图实现
type BucketHistogram struct {
	buckets     []float64          // 桶边界
	counts      []uint64           // 每个桶的计数
	sum         uint64             // 观测值总和（使用原子操作）
	totalCount  uint64             // 总观测次数
	labels      map[string]string
	mu          sync.RWMutex       // 保护buckets修改
}

// NewBucketHistogram 创建基于桶的直方图
func NewBucketHistogram(buckets []float64) Histogram {
	// 确保桶是有序的
	sortedBuckets := make([]float64, len(buckets))
	copy(sortedBuckets, buckets)
	sort.Float64s(sortedBuckets)
	
	return &BucketHistogram{
		buckets: sortedBuckets,
		counts:  make([]uint64, len(sortedBuckets)+1), // +1 for +Inf bucket
		labels:  make(map[string]string),
	}
}

// DefaultHistogramBuckets 默认的直方图桶
var DefaultHistogramBuckets = []float64{
	0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10,
}

// NewDefaultHistogram 创建使用默认桶的直方图
func NewDefaultHistogram() Histogram {
	return NewBucketHistogram(DefaultHistogramBuckets)
}

// Observe 记录观测值
func (h *BucketHistogram) Observe(value float64) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return // 忽略无效值
	}
	
	// 原子增加总计数和总和
	atomic.AddUint64(&h.totalCount, 1)
	atomic.AddUint64(&h.sum, floatToBits(value))
	
	// 找到对应的桶并增加计数
	h.mu.RLock()
	bucketIndex := h.findBucket(value)
	h.mu.RUnlock()
	
	atomic.AddUint64(&h.counts[bucketIndex], 1)
}

// findBucket 找到值应该放入的桶索引
func (h *BucketHistogram) findBucket(value float64) int {
	// 使用二分查找
	left, right := 0, len(h.buckets)
	for left < right {
		mid := (left + right) / 2
		if value <= h.buckets[mid] {
			right = mid
		} else {
			left = mid + 1
		}
	}
	return left
}

// Count 获取观测次数
func (h *BucketHistogram) Count() uint64 {
	return atomic.LoadUint64(&h.totalCount)
}

// Sum 获取观测值总和
func (h *BucketHistogram) Sum() float64 {
	bits := atomic.LoadUint64(&h.sum)
	return bitsToFloat(bits)
}

// Quantile 获取分位数（近似值）
func (h *BucketHistogram) Quantile(q float64) float64 {
	if q < 0 || q > 1 {
		return math.NaN()
	}
	
	totalCount := h.Count()
	if totalCount == 0 {
		return 0
	}
	
	targetCount := float64(totalCount) * q
	cumulativeCount := uint64(0)
	
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	for i, bucket := range h.buckets {
		cumulativeCount += atomic.LoadUint64(&h.counts[i])
		if float64(cumulativeCount) >= targetCount {
			return bucket
		}
	}
	
	return math.Inf(1) // +Inf
}

// Reset 重置直方图
func (h *BucketHistogram) Reset() {
	atomic.StoreUint64(&h.totalCount, 0)
	atomic.StoreUint64(&h.sum, 0)
	
	for i := range h.counts {
		atomic.StoreUint64(&h.counts[i], 0)
	}
}

// WithLabels 创建带标签的直方图副本
func (h *BucketHistogram) WithLabels(labels map[string]string) Histogram {
	newLabels := make(map[string]string, len(h.labels)+len(labels))
	for k, v := range h.labels {
		newLabels[k] = v
	}
	for k, v := range labels {
		newLabels[k] = v
	}
	
	h.mu.RLock()
	buckets := make([]float64, len(h.buckets))
	copy(buckets, h.buckets)
	h.mu.RUnlock()
	
	newHist := NewBucketHistogram(buckets).(*BucketHistogram)
	newHist.labels = newLabels
	return newHist
}

// GetBuckets 获取桶信息（用于调试和监控）
func (h *BucketHistogram) GetBuckets() []BucketInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	buckets := make([]BucketInfo, len(h.buckets)+1)
	
	for i, boundary := range h.buckets {
		buckets[i] = BucketInfo{
			UpperBound: boundary,
			Count:      atomic.LoadUint64(&h.counts[i]),
		}
	}
	
	// +Inf 桶
	buckets[len(h.buckets)] = BucketInfo{
		UpperBound: math.Inf(1),
		Count:      atomic.LoadUint64(&h.counts[len(h.buckets)]),
	}
	
	return buckets
}

// BucketInfo 已在interfaces.go中定义

// TimingHistogram 专门用于时间测量的直方图
type TimingHistogram struct {
	histogram Histogram
	unit      time.Duration // 时间单位
}

// NewTimingHistogram 创建时间直方图
func NewTimingHistogram(unit time.Duration) *TimingHistogram {
	// 为时间测量优化的桶（以毫秒为单位）
	timingBuckets := []float64{
		0.1, 0.5, 1, 2, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000,
	}
	
	return &TimingHistogram{
		histogram: NewBucketHistogram(timingBuckets),
		unit:      unit,
	}
}

// ObserveDuration 记录时间duration
func (th *TimingHistogram) ObserveDuration(duration time.Duration) {
	value := float64(duration) / float64(th.unit)
	th.histogram.Observe(value)
}

// ObserveMilliseconds 记录毫秒数
func (th *TimingHistogram) ObserveMilliseconds(ms float64) {
	th.histogram.Observe(ms)
}

// GetPercentile 获取百分位数（别名）
func (th *TimingHistogram) GetPercentile(p float64) time.Duration {
	value := th.histogram.Quantile(p / 100)
	return time.Duration(value * float64(th.unit))
}

// GetHistogram 获取底层直方图
func (th *TimingHistogram) GetHistogram() Histogram {
	return th.histogram
}

// HistogramTimer 直方图计时器，用于自动记录执行时间
type HistogramTimer struct {
	histogram *TimingHistogram
	startTime time.Time
}

// NewHistogramTimer 创建计时器
func NewHistogramTimer(histogram *TimingHistogram) *HistogramTimer {
	return &HistogramTimer{
		histogram: histogram,
		startTime: time.Now(),
	}
}

// ObserveDuration 记录从创建timer到现在的时间
func (t *HistogramTimer) ObserveDuration() time.Duration {
	duration := time.Since(t.startTime)
	t.histogram.ObserveDuration(duration)
	return duration
}

// SlidingWindowHistogram 滑动窗口直方图
type SlidingWindowHistogram struct {
	window       time.Duration
	buckets      []float64
	observations []observation
	mu           sync.RWMutex
}

type observation struct {
	value     float64
	timestamp time.Time
}

// NewSlidingWindowHistogram 创建滑动窗口直方图
func NewSlidingWindowHistogram(window time.Duration, buckets []float64) *SlidingWindowHistogram {
	sortedBuckets := make([]float64, len(buckets))
	copy(sortedBuckets, buckets)
	sort.Float64s(sortedBuckets)
	
	return &SlidingWindowHistogram{
		window:       window,
		buckets:      sortedBuckets,
		observations: make([]observation, 0),
	}
}

// Observe 记录观测值
func (swh *SlidingWindowHistogram) Observe(value float64) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return
	}
	
	now := time.Now()
	obs := observation{value: value, timestamp: now}
	
	swh.mu.Lock()
	// 添加新观测值
	swh.observations = append(swh.observations, obs)
	
	// 优化：只有当观测值数量较多时才清理，减少清理频率
	if len(swh.observations)%10 == 0 {
		swh.cleanOldObservations(now)
	}
	swh.mu.Unlock()
}

// cleanOldObservations 清理过期观测值
func (swh *SlidingWindowHistogram) cleanOldObservations(now time.Time) {
	cutoff := now.Add(-swh.window)
	
	// 找到第一个未过期的观测值
	start := 0
	for i, obs := range swh.observations {
		if obs.timestamp.After(cutoff) {
			start = i
			break
		}
	}
	
	// 如果有过期的观测值，移除它们
	if start > 0 {
		swh.observations = swh.observations[start:]
	}
}

// Count 获取窗口内观测次数
func (swh *SlidingWindowHistogram) Count() uint64 {
	swh.mu.RLock()
	defer swh.mu.RUnlock()
	
	swh.cleanOldObservations(time.Now())
	return uint64(len(swh.observations))
}

// Sum 获取窗口内观测值总和
func (swh *SlidingWindowHistogram) Sum() float64 {
	swh.mu.RLock()
	defer swh.mu.RUnlock()
	
	swh.cleanOldObservations(time.Now())
	
	var sum float64
	for _, obs := range swh.observations {
		sum += obs.value
	}
	return sum
}

// Quantile 获取窗口内分位数
func (swh *SlidingWindowHistogram) Quantile(q float64) float64 {
	if q < 0 || q > 1 {
		return math.NaN()
	}
	
	swh.mu.RLock()
	defer swh.mu.RUnlock()
	
	swh.cleanOldObservations(time.Now())
	
	if len(swh.observations) == 0 {
		return 0
	}
	
	// 提取值并排序
	values := make([]float64, len(swh.observations))
	for i, obs := range swh.observations {
		values[i] = obs.value
	}
	sort.Float64s(values)
	
	// 计算分位数
	index := q * float64(len(values)-1)
	lower := int(index)
	upper := lower + 1
	
	if upper >= len(values) {
		return values[lower]
	}
	
	// 线性插值
	weight := index - float64(lower)
	return values[lower]*(1-weight) + values[upper]*weight
}

// Reset 重置滑动窗口直方图
func (swh *SlidingWindowHistogram) Reset() {
	swh.mu.Lock()
	defer swh.mu.Unlock()
	
	swh.observations = swh.observations[:0]
}

// WithLabels 创建带标签的滑动窗口直方图
func (swh *SlidingWindowHistogram) WithLabels(labels map[string]string) Histogram {
	return NewSlidingWindowHistogram(swh.window, swh.buckets)
}

// HistogramPerformanceTracker 性能追踪器
type HistogramPerformanceTracker struct {
	histograms map[string]*TimingHistogram
	mu         sync.RWMutex
}

// NewHistogramPerformanceTracker 创建性能追踪器
func NewHistogramPerformanceTracker() *HistogramPerformanceTracker {
	return &HistogramPerformanceTracker{
		histograms: make(map[string]*TimingHistogram),
	}
}

// Track 追踪操作性能
func (pt *HistogramPerformanceTracker) Track(name string) func() {
	start := time.Now()
	
	return func() {
		duration := time.Since(start)
		pt.RecordDuration(name, duration)
	}
}

// RecordDuration 记录操作耗时
func (pt *HistogramPerformanceTracker) RecordDuration(name string, duration time.Duration) {
	hist := pt.getOrCreateHistogram(name)
	hist.ObserveDuration(duration)
}

// getOrCreateHistogram 获取或创建直方图
func (pt *HistogramPerformanceTracker) getOrCreateHistogram(name string) *TimingHistogram {
	pt.mu.RLock()
	hist, exists := pt.histograms[name]
	pt.mu.RUnlock()
	
	if exists {
		return hist
	}
	
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	// 双检查锁定
	if hist, exists := pt.histograms[name]; exists {
		return hist
	}
	
	hist = NewTimingHistogram(time.Millisecond)
	pt.histograms[name] = hist
	return hist
}

// GetStats 获取性能统计
func (pt *HistogramPerformanceTracker) GetStats(name string) *HistogramPerformanceStats {
	pt.mu.RLock()
	hist, exists := pt.histograms[name]
	pt.mu.RUnlock()
	
	if !exists {
		return nil
	}
	
	h := hist.GetHistogram()
	return &HistogramPerformanceStats{
		Name:        name,
		Count:       h.Count(),
		Sum:         time.Duration(h.Sum()) * time.Millisecond,
		P50:         hist.GetPercentile(50),
		P95:         hist.GetPercentile(95),
		P99:         hist.GetPercentile(99),
	}
}

// GetAllStats 获取所有性能统计
func (pt *HistogramPerformanceTracker) GetAllStats() map[string]*HistogramPerformanceStats {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	
	stats := make(map[string]*HistogramPerformanceStats, len(pt.histograms))
	for name := range pt.histograms {
		stats[name] = pt.GetStats(name)
	}
	return stats
}

// HistogramPerformanceStats 性能统计信息
type HistogramPerformanceStats struct {
	Name  string        `json:"name"`
	Count uint64        `json:"count"`
	Sum   time.Duration `json:"sum"`
	P50   time.Duration `json:"p50"`
	P95   time.Duration `json:"p95"`
	P99   time.Duration `json:"p99"`
}

// Average 获取平均耗时
func (ps *HistogramPerformanceStats) Average() time.Duration {
	if ps.Count == 0 {
		return 0
	}
	return ps.Sum / time.Duration(ps.Count)
}

// LabeledHistogram 带标签的直方图管理器
type LabeledHistogram struct {
	mu         sync.RWMutex
	histograms map[string]Histogram
	factory    func() Histogram
}

// NewLabeledHistogram 创建标签直方图管理器
func NewLabeledHistogram() *LabeledHistogram {
	return &LabeledHistogram{
		histograms: make(map[string]Histogram),
		factory:    NewDefaultHistogram,
	}
}

// WithLabelValues 根据标签值获取直方图
func (lh *LabeledHistogram) WithLabelValues(labelValues ...string) Histogram {
	key := generateKey(labelValues...)
	
	lh.mu.RLock()
	histogram, exists := lh.histograms[key]
	lh.mu.RUnlock()
	
	if exists {
		return histogram
	}
	
	lh.mu.Lock()
	defer lh.mu.Unlock()
	
	// 双检查锁定模式
	if histogram, exists := lh.histograms[key]; exists {
		return histogram
	}
	
	histogram = lh.factory()
	lh.histograms[key] = histogram
	return histogram
}

// Reset 重置所有直方图
func (lh *LabeledHistogram) Reset() {
	lh.mu.Lock()
	defer lh.mu.Unlock()
	
	for _, histogram := range lh.histograms {
		histogram.Reset()
	}
}

// GetAllStats 获取所有直方图统计
func (lh *LabeledHistogram) GetAllStats() map[string]BucketHistogramStats {
	lh.mu.RLock()
	defer lh.mu.RUnlock()
	
	stats := make(map[string]BucketHistogramStats, len(lh.histograms))
	for key, histogram := range lh.histograms {
		stats[key] = BucketHistogramStats{
			Count: histogram.Count(),
			Sum:   histogram.Sum(),
			P50:   histogram.Quantile(0.5),
			P95:   histogram.Quantile(0.95),
			P99:   histogram.Quantile(0.99),
		}
	}
	
	return stats
}

// BucketHistogramStats 直方图统计信息
type BucketHistogramStats struct {
	Count uint64  `json:"count"`
	Sum   float64 `json:"sum"`
	P50   float64 `json:"p50"`
	P95   float64 `json:"p95"`
	P99   float64 `json:"p99"`
}

// Average 获取平均值
func (hs *BucketHistogramStats) Average() float64 {
	if hs.Count == 0 {
		return 0
	}
	return hs.Sum / float64(hs.Count)
}