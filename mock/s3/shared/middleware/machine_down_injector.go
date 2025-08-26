package middleware

import (
	"context"
	"mocks3/shared/observability"
	"runtime"
	"sync"
	"time"
)

// MachineDownInjector 机器宕机异常注入器
type MachineDownInjector struct {
	logger   *observability.Logger
	isActive bool
	mu       sync.RWMutex
	stopChan chan struct{}
}

// NewMachineDownInjector 创建机器宕机异常注入器
func NewMachineDownInjector(logger *observability.Logger) *MachineDownInjector {
	return &MachineDownInjector{
		logger:   logger,
		stopChan: make(chan struct{}),
	}
}

// StartMachineDown 开始机器宕机异常注入
// 此方法会模拟机器故障，包括服务停止和资源耗尽
func (m *MachineDownInjector) StartMachineDown(ctx context.Context, simulationType string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isActive {
		m.logger.Warn(ctx, "Machine down injection already active")
		return
	}

	m.isActive = true
	m.logger.Info(ctx, "Starting machine down injection",
		observability.String("simulation_type", simulationType),
		observability.String("duration", duration.String()))

	// 根据模拟类型启动不同的故障模拟
	switch simulationType {
	case "service_hang":
		go m.serviceHangTask(ctx)
	case "resource_exhaustion":
		go m.resourceExhaustionTask(ctx)
	case "network_isolation":
		go m.networkIsolationTask(ctx)
	default:
		// 默认使用服务挂起模拟
		go m.serviceHangTask(ctx)
	}

	// 设置定时器自动停止
	go func() {
		select {
		case <-time.After(duration):
			m.StopMachineDown(ctx)
		case <-m.stopChan:
			return
		}
	}()
}

// StopMachineDown 停止机器宕机异常注入
func (m *MachineDownInjector) StopMachineDown(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isActive {
		return
	}

	m.logger.Info(ctx, "Stopping machine down injection")
	m.isActive = false
}

// IsActive 检查机器宕机注入是否活跃
func (m *MachineDownInjector) IsActive() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isActive
}

// serviceHangTask 服务挂起任务 - 通过阻塞主要协程模拟服务无响应
func (m *MachineDownInjector) serviceHangTask(ctx context.Context) {
	m.logger.Warn(ctx, "Service hang simulation started - blocking operations")

	// 创建多个阻塞协程，消耗系统资源
	for i := 0; i < runtime.NumCPU()*2; i++ {
		go func(workerID int) {
			for {
				select {
				case <-m.stopChan:
					return
				default:
					// 模拟长时间阻塞操作
					time.Sleep(10 * time.Second)
				}
			}
		}(i)
	}

	// 主协程也进入阻塞状态
	for {
		select {
		case <-m.stopChan:
			m.logger.Info(ctx, "Service hang simulation stopped")
			return
		default:
			time.Sleep(5 * time.Second)
		}
	}
}

// resourceExhaustionTask 资源耗尽任务 - 通过快速消耗系统资源模拟机器故障
func (m *MachineDownInjector) resourceExhaustionTask(ctx context.Context) {
	m.logger.Warn(ctx, "Resource exhaustion simulation started")

	// 创建大量协程消耗CPU
	for i := 0; i < runtime.NumCPU()*10; i++ {
		go func() {
			for {
				select {
				case <-m.stopChan:
					return
				default:
					// 高强度CPU计算
					for j := 0; j < 1000000; j++ {
						_ = j * j * j
					}
				}
			}
		}()
	}

	// 快速分配内存
	go func() {
		var memoryChunks [][]byte
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-m.stopChan:
				// 清理内存
				memoryChunks = nil
				runtime.GC()
				return
			case <-ticker.C:
				// 每次分配50MB
				chunk := make([]byte, 50*1024*1024)
				// 写入数据确保内存真正被使用
				for i := range chunk {
					if i%1024 == 0 {
						chunk[i] = byte(i % 256)
					}
				}
				memoryChunks = append(memoryChunks, chunk)

				m.logger.Debug(ctx, "Memory allocated for exhaustion simulation",
					observability.Int("total_chunks", len(memoryChunks)))
			}
		}
	}()

	// 等待停止信号
	<-m.stopChan
	m.logger.Info(ctx, "Resource exhaustion simulation stopped")
}

// networkIsolationTask 网络隔离任务 - 通过阻塞网络操作模拟网络故障
func (m *MachineDownInjector) networkIsolationTask(ctx context.Context) {
	m.logger.Warn(ctx, "Network isolation simulation started")

	// 在Linux/Unix系统上，可以尝试修改网络配置（需要适当权限）
	if runtime.GOOS == "linux" {
		// 注意：这些命令需要root权限，在生产环境中要谨慎使用
		m.logger.Info(ctx, "Attempting network isolation (requires elevated privileges)")

		// 这里仅作为示例，实际使用时需要根据具体环境调整
		// 可以通过iptables规则、网络namespace等方式实现网络隔离
		// 为了安全性，这里不实际执行系统命令
	}

	// 模拟网络超时和连接失败
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopChan:
			m.logger.Info(ctx, "Network isolation simulation stopped")
			return
		case <-ticker.C:
			m.logger.Debug(ctx, "Simulating network isolation - blocking network operations")
			// 这里可以添加网络阻塞逻辑
		}
	}
}

// Cleanup 清理资源
func (m *MachineDownInjector) Cleanup() {
	close(m.stopChan)
	m.StopMachineDown(context.Background())
}
