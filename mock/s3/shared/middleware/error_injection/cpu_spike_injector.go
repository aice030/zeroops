package error_injection

import (
	"context"
	"mocks3/shared/observability"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// CPUSpikeInjector CPU峰值异常注入器
type CPUSpikeInjector struct {
	logger        *observability.Logger
	isActive      bool
	mu            sync.RWMutex
	stopChan      chan struct{}
	goroutines    []chan struct{}
	targetPercent float64 // 目标CPU使用率
	// CPU统计
	lastTotal  uint64
	lastIdle   uint64
	lastUpdate time.Time
}

// NewCPUSpikeInjector 创建CPU峰值异常注入器
func NewCPUSpikeInjector(logger *observability.Logger) *CPUSpikeInjector {
	return &CPUSpikeInjector{
		logger:   logger,
		stopChan: make(chan struct{}),
	}
}

// StartCPUSpike 开始CPU峰值异常注入
func (c *CPUSpikeInjector) StartCPUSpike(ctx context.Context, targetCPUPercent float64, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isActive {
		c.logger.Warn(ctx, "CPU spike injection already active")
		return
	}

	c.isActive = true
	c.targetPercent = targetCPUPercent
	c.logger.Info(ctx, "Starting CPU spike injection",
		observability.Float64("target_cpu_percent", targetCPUPercent),
		observability.String("duration", duration.String()))

	// 计算需要的协程数量
	numCPU := runtime.NumCPU()
	numGoroutines := int(float64(numCPU) * targetCPUPercent / 100.0)
	if numGoroutines < 1 {
		numGoroutines = 1
	}

	// 启动CPU密集型协程
	for i := 0; i < numGoroutines; i++ {
		stopChan := make(chan struct{})
		c.goroutines = append(c.goroutines, stopChan)
		go c.cpuIntensiveTask(stopChan)
	}

	// 设置定时器自动停止
	go func() {
		select {
		case <-time.After(duration):
			c.StopCPUSpike(ctx)
		case <-c.stopChan:
			return
		}
	}()
}

// StopCPUSpike 停止CPU峰值异常注入
func (c *CPUSpikeInjector) StopCPUSpike(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isActive {
		return
	}

	c.logger.Info(ctx, "Stopping CPU spike injection")
	c.isActive = false
	c.targetPercent = 0

	// 停止所有CPU密集型协程
	for _, stopChan := range c.goroutines {
		close(stopChan)
	}
	c.goroutines = nil
}

// IsActive 检查CPU峰值注入是否活跃
func (c *CPUSpikeInjector) IsActive() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isActive
}

// GetCurrentCPUUsage 获取当前CPU使用率
func (c *CPUSpikeInjector) GetCurrentCPUUsage() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.isActive {
		return c.readSystemCPUUsage()
	}

	// CPU注入活跃时，读取真实的CPU使用率
	return c.readSystemCPUUsage()
}

// readSystemCPUUsage 读取系统真实CPU使用率
func (c *CPUSpikeInjector) readSystemCPUUsage() float64 {
	// 读取/proc/stat获取CPU统计
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0.0
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		return 0.0
	}

	// 解析第一行 CPU总计
	fields := strings.Fields(lines[0])
	if len(fields) < 8 || fields[0] != "cpu" {
		return 0.0
	}

	// 解析CPU时间值
	var values [7]uint64
	for i := 0; i < 7; i++ {
		val, err := strconv.ParseUint(fields[i+1], 10, 64)
		if err != nil {
			return 0.0
		}
		values[i] = val
	}

	// 计算总时间和空闲时间
	total := values[0] + values[1] + values[2] + values[3] + values[4] + values[5] + values[6]
	idle := values[3] + values[4] // idle + iowait

	now := time.Now()

	// 第一次读取，保存基准值
	if c.lastTotal == 0 {
		c.lastTotal = total
		c.lastIdle = idle
		c.lastUpdate = now
		return 0.0
	}

	// 计算时间差值
	totalDiff := total - c.lastTotal
	idleDiff := idle - c.lastIdle

	// 更新基准值
	c.lastTotal = total
	c.lastIdle = idle
	c.lastUpdate = now

	// 计算CPU使用率
	if totalDiff > 0 {
		cpuUsage := float64(totalDiff-idleDiff) / float64(totalDiff) * 100.0
		if cpuUsage < 0 {
			cpuUsage = 0
		}
		if cpuUsage > 100 {
			cpuUsage = 100
		}
		return cpuUsage
	}

	return 0.0
}

// cpuIntensiveTask CPU密集型任务
func (c *CPUSpikeInjector) cpuIntensiveTask(stopChan chan struct{}) {
	for {
		select {
		case <-stopChan:
			return
		default:
			// 执行CPU密集型计算
			for i := 0; i < 10000; i++ {
				_ = i * i * i
			}
			// 短暂让出CPU，避免完全阻塞
			runtime.Gosched()
		}
	}
}

// Cleanup 清理资源
func (c *CPUSpikeInjector) Cleanup() {
	close(c.stopChan)
	c.StopCPUSpike(context.Background())
}
