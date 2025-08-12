package logs

import (
	"context"
	"shared/config"
)

// Logger 简化的日志接口
type Logger interface {
	Info(ctx context.Context, message string, fields map[string]any)
	Error(ctx context.Context, message string, err error, fields map[string]any)
	Debug(ctx context.Context, message string, fields map[string]any)
	Warn(ctx context.Context, message string, fields map[string]any)
}

// NewLogger 创建日志器
func NewLogger(config config.LoggingConfig) Logger {
	// TODO: 实现日志器
	return nil
}
