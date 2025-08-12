package httpserver

import (
	"context"
	"net/http"
	"shared/config"
)

// Server HTTP服务器接口
type Server interface {
	// Start 启动服务器
	Start() error

	// Stop 停止服务器
	Stop(ctx context.Context) error

	// AddHandler 添加路由处理器
	AddHandler(pattern string, handler http.Handler)

	// AddHandlerFunc 添加路由处理函数
	AddHandlerFunc(pattern string, handler http.HandlerFunc)

	// GetAddr 获取服务器监听地址
	GetAddr() string

	// IsRunning 检查服务器是否在运行
	IsRunning() bool
}

// NewServer 创建HTTP服务器
func NewServer(config config.HTTPServerConfig) Server {
	// TODO: 实现HTTP服务器
	return nil
}
