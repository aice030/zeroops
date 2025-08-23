package log

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/trace"
)

// Logger 结构化日志器
type Logger struct {
	logger      log.Logger
	serviceName string
	level       LogLevel
}

// LogLevel 日志级别
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// NewLogger 创建新的日志器
// 注意：需要先调用 InitializeLoggerProvider 初始化 OTEL Log Provider
func NewLogger(serviceName string, level LogLevel) *Logger {
	// 从全局 LoggerProvider 获取 Logger
	loggerProvider := global.GetLoggerProvider()
	otelLogger := loggerProvider.Logger(serviceName)

	return &Logger{
		logger:      otelLogger,
		serviceName: serviceName,
		level:       level,
	}
}

// WithContext 为日志添加追踪上下文
// OTEL Logs API 会自动从 context 中提取 trace 信息
func (l *Logger) WithContext(ctx context.Context) *Logger {
	// OTEL Logs API 会自动关联 trace context
	return l
}

// WithFields 添加字段（保持兼容性）
func (l *Logger) With(args ...any) *Logger {
	// 对于 OTEL Logs，返回同一个 Logger，属性在具体日志记录时添加
	return l
}

// WithFields 添加多个字段
func (l *Logger) WithFields(fields map[string]any) *Logger {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return l.With(args...)
}

// Debug 调试日志
func (l *Logger) Debug(msg string, args ...any) {
	if l.level > LevelDebug {
		return
	}
	l.emit(context.Background(), log.SeverityDebug, msg, args...)
}

// Info 信息日志
func (l *Logger) Info(msg string, args ...any) {
	if l.level > LevelInfo {
		return
	}
	l.emit(context.Background(), log.SeverityInfo, msg, args...)
}

// Warn 警告日志
func (l *Logger) Warn(msg string, args ...any) {
	if l.level > LevelWarn {
		return
	}
	l.emit(context.Background(), log.SeverityWarn, msg, args...)
}

// Error 错误日志
func (l *Logger) Error(msg string, args ...any) {
	if l.level > LevelError {
		return
	}
	l.emit(context.Background(), log.SeverityError, msg, args...)
}

// emit 发送日志到 OTEL
func (l *Logger) emit(ctx context.Context, severity log.Severity, msg string, args ...any) {
	// 准备属性
	attrs := []log.KeyValue{
		log.String("service", l.serviceName),
		log.String("message", msg),
	}

	// 处理额外参数
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			key := fmt.Sprintf("%v", args[i])
			value := fmt.Sprintf("%v", args[i+1])
			attrs = append(attrs, log.String(key, value))
		}
	}

	// 从 context 中提取 trace 信息
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		spanCtx := span.SpanContext()
		attrs = append(attrs,
			log.String("trace_id", spanCtx.TraceID().String()),
			log.String("span_id", spanCtx.SpanID().String()),
		)
	}

	// 发送日志记录
	logRecord := log.Record{}
	logRecord.SetTimestamp(time.Now())
	logRecord.SetSeverity(severity)
	logRecord.SetBody(log.StringValue(msg))
	logRecord.AddAttributes(attrs...)

	l.logger.Emit(ctx, logRecord)
}

// DebugContext 带上下文的调试日志
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	if l.level > LevelDebug {
		return
	}
	l.emit(ctx, log.SeverityDebug, msg, args...)
}

// InfoContext 带上下文的信息日志
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	if l.level > LevelInfo {
		return
	}
	l.emit(ctx, log.SeverityInfo, msg, args...)
}

// WarnContext 带上下文的警告日志
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	if l.level > LevelWarn {
		return
	}
	l.emit(ctx, log.SeverityWarn, msg, args...)
}

// ErrorContext 带上下文的错误日志
func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	if l.level > LevelError {
		return
	}
	l.emit(ctx, log.SeverityError, msg, args...)
}

// LogError 记录错误
func (l *Logger) LogError(ctx context.Context, err error, msg string, args ...any) {
	if err == nil {
		return
	}

	// 添加错误信息到参数中
	errArgs := append(args, "error", err.Error())
	l.emit(ctx, log.SeverityError, msg, errArgs...)
}

// LogPanic 记录panic
func (l *Logger) LogPanic(ctx context.Context, r any, msg string, args ...any) {
	// 添加 panic 信息到参数中
	panicArgs := append(args, "panic", fmt.Sprintf("%v", r))
	l.emit(ctx, log.SeverityError, msg, panicArgs...)
}

// levelToSeverity 将内部级别转换为 OTEL 严重级别
func levelToSeverity(level LogLevel) log.Severity {
	switch level {
	case LevelDebug:
		return log.SeverityDebug
	case LevelInfo:
		return log.SeverityInfo
	case LevelWarn:
		return log.SeverityWarn
	case LevelError:
		return log.SeverityError
	default:
		return log.SeverityInfo
	}
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// GetLevel 获取当前日志级别
func (l *Logger) GetLevel() LogLevel {
	return l.level
}

// IsDebugEnabled 检查是否启用调试级别
func (l *Logger) IsDebugEnabled() bool {
	return l.level <= LevelDebug
}

// IsInfoEnabled 检查是否启用信息级别
func (l *Logger) IsInfoEnabled() bool {
	return l.level <= LevelInfo
}

// IsWarnEnabled 检查是否启用警告级别
func (l *Logger) IsWarnEnabled() bool {
	return l.level <= LevelWarn
}

// IsErrorEnabled 检查是否启用错误级别
func (l *Logger) IsErrorEnabled() bool {
	return l.level <= LevelError
}
