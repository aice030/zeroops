package logs

import (
	"context"
)

// LogCollector  定义日志记录器接口
type LogCollector interface {
	// Info 记录信息级别日志
	Info(ctx context.Context, message string, fields map[string]interface{})

	// Error 记录错误级别日志
	Error(ctx context.Context, message string, err error, fields map[string]interface{})

	// Debug 记录调试级别日志
	Debug(ctx context.Context, message string, fields map[string]interface{})

	// Warn 记录警告级别日志
	Warn(ctx context.Context, message string, fields map[string]interface{})
}

// LogEntry 结构化日志对象
type LogEntry struct {
	Timestamp int64                  // 时间戳，UnixNano
	Level     string                 // 日志级别，如 info, error
	Message   string                 // 日志内容
	Fields    map[string]interface{} // 额外字段，如 trace_id、请求ID等
}

// LogExporter 负责将日志数据推送到 Elasticsearch 或其他日志存储系统
type LogExporter interface {
	// Export 发送单条日志数据，异步或同步实现均可
	Export(ctx context.Context, entry LogEntry) error

	// Flush 推送缓冲中的日志（如果有缓存的话）
	Flush(ctx context.Context) error

	// Close 关闭导出器，释放资源
	Close() error
}
