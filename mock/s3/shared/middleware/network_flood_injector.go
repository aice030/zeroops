package middleware

import (
	"context"
	"mocks3/shared/observability"
	"net"
	"sync"
	"time"
)

// NetworkFloodInjector 网络风暴异常注入器
type NetworkFloodInjector struct {
	logger       *observability.Logger
	isActive     bool
	mu           sync.RWMutex
	stopChan     chan struct{}
	connections  []net.Conn
	targetConns  int
	currentConns int
}

// NewNetworkFloodInjector 创建网络风暴异常注入器
func NewNetworkFloodInjector(logger *observability.Logger) *NetworkFloodInjector {
	return &NetworkFloodInjector{
		logger:   logger,
		stopChan: make(chan struct{}),
	}
}

// StartNetworkFlood 开始网络风暴异常注入
func (n *NetworkFloodInjector) StartNetworkFlood(ctx context.Context, targetConnections int, duration time.Duration) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.isActive {
		n.logger.Warn(ctx, "Network flood injection already active")
		return
	}

	n.isActive = true
	n.targetConns = targetConnections
	n.logger.Info(ctx, "Starting network flood injection",
		observability.Int("target_connections", targetConnections),
		observability.String("duration", duration.String()))

	// 启动网络连接创建协程
	go n.networkFloodTask(ctx)

	// 设置定时器自动停止
	go func() {
		select {
		case <-time.After(duration):
			n.StopNetworkFlood(ctx)
		case <-n.stopChan:
			return
		}
	}()
}

// StopNetworkFlood 停止网络风暴异常注入
func (n *NetworkFloodInjector) StopNetworkFlood(ctx context.Context) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.isActive {
		return
	}

	n.logger.Info(ctx, "Stopping network flood injection",
		observability.Int("created_connections", n.currentConns))
	n.isActive = false

	// 关闭所有创建的连接
	for _, conn := range n.connections {
		if conn != nil {
			if err := conn.Close(); err != nil {
				n.logger.Warn(ctx, "Failed to close connection",
					observability.Error(err))
			}
		}
	}
	n.connections = nil
	n.currentConns = 0
}

// IsActive 检查网络风暴注入是否活跃
func (n *NetworkFloodInjector) IsActive() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.isActive
}

// GetCurrentConnections 获取当前创建的连接数
func (n *NetworkFloodInjector) GetCurrentConnections() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.currentConns
}

// networkFloodTask 网络风暴任务
func (n *NetworkFloodInjector) networkFloodTask(ctx context.Context) {
	ticker := time.NewTicker(100 * time.Millisecond) // 每100毫秒创建连接
	defer ticker.Stop()

	// 目标地址列表 - 使用一些公共的但响应较慢的服务
	targets := []string{
		"8.8.8.8:53",        // Google DNS
		"1.1.1.1:53",        // Cloudflare DNS
		"208.67.222.222:53", // OpenDNS
		"9.9.9.9:53",        // Quad9 DNS
	}

	targetIndex := 0

	for {
		select {
		case <-n.stopChan:
			return
		case <-ticker.C:
			n.mu.Lock()
			if !n.isActive {
				n.mu.Unlock()
				return
			}

			// 检查是否达到目标连接数
			if n.currentConns >= n.targetConns {
				n.mu.Unlock()
				continue
			}

			// 选择目标地址
			target := targets[targetIndex%len(targets)]
			targetIndex++

			// 创建连接
			dialer := net.Dialer{
				Timeout: 2 * time.Second,
			}

			conn, err := dialer.DialContext(ctx, "tcp", target)
			if err != nil {
				n.logger.Debug(ctx, "Failed to create network connection",
					observability.String("target", target),
					observability.Error(err))
				n.mu.Unlock()
				continue
			}

			n.connections = append(n.connections, conn)
			n.currentConns++

			n.logger.Info(ctx, "Network connection created",
				observability.String("target", target),
				observability.Int("total_connections", n.currentConns),
				observability.Int("target_connections", n.targetConns))

			// 在后台发送一些数据以保持连接活跃
			go n.keepConnectionActive(conn)

			n.mu.Unlock()
		}
	}
}

// keepConnectionActive 保持连接活跃
func (n *NetworkFloodInjector) keepConnectionActive(conn net.Conn) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	defer conn.Close()

	// 创建一些测试数据
	testData := []byte("test data for network flood simulation\n")

	for {
		select {
		case <-n.stopChan:
			return
		case <-ticker.C:
			// 设置写入超时
			conn.SetWriteDeadline(time.Now().Add(1 * time.Second))

			// 发送测试数据
			if _, err := conn.Write(testData); err != nil {
				// 连接已断开，退出
				return
			}

			// 尝试读取响应（如果有的话）
			conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			buffer := make([]byte, 1024)
			_, _ = conn.Read(buffer) // 忽略错误，因为某些服务可能不响应
		}
	}
}

// Cleanup 清理资源
func (n *NetworkFloodInjector) Cleanup() {
	close(n.stopChan)
	n.StopNetworkFlood(context.Background())
}
