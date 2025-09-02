package error_injection

import (
	"context"
	"mocks3/shared/observability"
	"runtime"
	"sync"
	"time"
)

// MemoryLeakInjector 内存泄露异常注入器
type MemoryLeakInjector struct {
	logger     *observability.Logger
	isActive   bool
	mu         sync.RWMutex
	stopChan   chan struct{}
	memoryPool [][]byte
	targetMB   int64
	currentMB  int64
}

// NewMemoryLeakInjector 创建内存泄露异常注入器
func NewMemoryLeakInjector(logger *observability.Logger) *MemoryLeakInjector {
	return &MemoryLeakInjector{
		logger:   logger,
		stopChan: make(chan struct{}),
	}
}

// StartMemoryLeak 开始内存泄露异常注入
func (m *MemoryLeakInjector) StartMemoryLeak(ctx context.Context, targetMemoryMB int64, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isActive {
		m.logger.Warn(ctx, "Memory leak injection already active")
		return
	}

	m.isActive = true
	m.targetMB = targetMemoryMB
	m.logger.Info(ctx, "Starting memory leak injection",
		observability.Int64("target_memory_mb", targetMemoryMB),
		observability.String("duration", duration.String()))

	// 启动内存分配协程
	go m.memoryAllocationTask(ctx)

	// 设置定时器自动停止
	go func() {
		select {
		case <-time.After(duration):
			m.StopMemoryLeak(ctx)
		case <-m.stopChan:
			return
		}
	}()
}

// StopMemoryLeak 停止内存泄露异常注入
func (m *MemoryLeakInjector) StopMemoryLeak(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isActive {
		return
	}

	m.logger.Info(ctx, "Stopping memory leak injection",
		observability.Int64("allocated_memory_mb", m.currentMB))
	m.isActive = false

	// 释放所有分配的内存
	m.memoryPool = nil
	m.currentMB = 0

	// 强制垃圾回收
	runtime.GC()
}

// IsActive 检查内存泄露注入是否活跃
func (m *MemoryLeakInjector) IsActive() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isActive
}

// GetCurrentMemoryMB 获取当前分配的内存大小（MB）
func (m *MemoryLeakInjector) GetCurrentMemoryMB() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentMB
}

// memoryAllocationTask 内存分配任务
func (m *MemoryLeakInjector) memoryAllocationTask(ctx context.Context) {
	// 使用更长的分配间隔，减少GC压力
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.mu.Lock()
			if !m.isActive {
				m.mu.Unlock()
				return
			}

			// 检查是否达到目标内存
			if m.currentMB >= m.targetMB {
				m.mu.Unlock()
				continue
			}

			// 增加每次分配的块大小，减少分配频率
			chunkSizeMB := int64(50) // 从10MB增加到50MB
			if m.currentMB+chunkSizeMB > m.targetMB {
				chunkSizeMB = m.targetMB - m.currentMB
			}

			// 分配内存（1MB = 1024*1024 bytes）
			chunk := make([]byte, chunkSizeMB*1024*1024)

			// 优化内存写入策略，减少GC触发
			// 使用更稀疏的写入模式，减少内存访问压力
			for i := 0; i < len(chunk); i += 4096 { // 每4KB写入一次
				chunk[i] = byte(i % 256)
			}

			m.memoryPool = append(m.memoryPool, chunk)
			m.currentMB += chunkSizeMB

			m.logger.Info(ctx, "Memory allocated",
				observability.Int64("allocated_mb", chunkSizeMB),
				observability.Int64("total_allocated_mb", m.currentMB),
				observability.Int64("target_mb", m.targetMB))

			m.mu.Unlock()

			// 添加短暂延迟，让系统稳定
			time.Sleep(500 * time.Millisecond)
		}
	}
}

// Cleanup 清理资源
func (m *MemoryLeakInjector) Cleanup() {
	close(m.stopChan)
	m.StopMemoryLeak(context.Background())
}
