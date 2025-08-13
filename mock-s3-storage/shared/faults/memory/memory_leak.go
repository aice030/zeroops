package memory

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"shared/faults"
)

type MemLeakFault struct {
	running    int32         // 标志故障是否运行中
	stopCh     chan struct{} // 用于停止 goroutine
	stoppedWg  sync.WaitGroup
	leakData   [][]byte      // 存储分配的内存，防止 GC 回收
	allocSize  int           // 每次分配的内存大小（字节）
	allocDelay time.Duration // 分配间隔
}

func NewMemLeakFault(allocSize int, allocDelay time.Duration) *MemLeakFault {
	return &MemLeakFault{
		stopCh:     make(chan struct{}),
		allocSize:  allocSize,
		allocDelay: allocDelay,
	}
}

func (m *MemLeakFault) Name() string {
	return "MemLeak"
}

func (m *MemLeakFault) Start() error {
	// 首先判断运行状态，防止重复调用
	if !atomic.CompareAndSwapInt32(&m.running, 0, 1) {
		return fmt.Errorf("MemLeakFault already running")
	}
	m.stopCh = make(chan struct{})
	m.stoppedWg.Add(1)

	go func() {
		defer m.stoppedWg.Done()
		ticker := time.NewTicker(m.allocDelay)
		defer ticker.Stop()

		for {
			select {
			case <-m.stopCh:
				return
			case <-ticker.C:
				// 分配内存并存储
				block := make([]byte, m.allocSize)
				for i := range block {
					block[i] = byte(i % 256) // 写入数据，避免惰性分配
				}
				m.leakData = append(m.leakData, block)
			}
		}
	}()
	return nil
}

func (m *MemLeakFault) Stop() error {
	if !atomic.CompareAndSwapInt32(&m.running, 1, 0) {
		return fmt.Errorf("MemLeakFault not running")
	}
	close(m.stopCh)
	m.stoppedWg.Wait()
	m.leakData = nil // 释放内存引用，方便GC回收
	return nil
}

func (m *MemLeakFault) Status() string {
	if atomic.LoadInt32(&m.running) == 1 {
		return "running"
	}
	return "stopped"
}

// 确保 MemLeakFault 实现了 faults.Fault 接口
var _ faults.Fault = (*MemLeakFault)(nil)
