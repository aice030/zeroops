package middleware

import (
	"context"
	"mocks3/shared/observability"
	"runtime"
	"sync"
	"time"
)

// CPUSpikeInjector CPU峰值异常注入器
type CPUSpikeInjector struct {
	logger    *observability.Logger
	isActive  bool
	mu        sync.RWMutex
	stopChan  chan struct{}
	goroutines []chan struct{}
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