package middleware

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
	ticker := time.NewTicker(1 * time.Second) // 每秒分配一次
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

			// 每次分配10MB
			chunkSizeMB := int64(10)
			if m.currentMB+chunkSizeMB > m.targetMB {
				chunkSizeMB = m.targetMB - m.currentMB
			}

			// 分配内存（1MB = 1024*1024 bytes）
			chunk := make([]byte, chunkSizeMB*1024*1024)

			// 写入数据确保内存真正被使用
			for i := range chunk {
				if i%1024 == 0 {
					chunk[i] = byte(i % 256)
				}
			}

			m.memoryPool = append(m.memoryPool, chunk)
			m.currentMB += chunkSizeMB

			m.logger.Info(ctx, "Memory allocated",
				observability.Int64("allocated_mb", chunkSizeMB),
				observability.Int64("total_allocated_mb", m.currentMB),
				observability.Int64("target_mb", m.targetMB))

			m.mu.Unlock()
		}
	}
}

// Cleanup 清理资源
func (m *MemoryLeakInjector) Cleanup() {
	close(m.stopChan)
	m.StopMemoryLeak(context.Background())
}
