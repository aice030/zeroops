package observability

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/trace"
)

// LogLevel 日志级别
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

// Field 日志字段
type Field struct {
	Key   string
	Value any
}

// String 创建字符串字段
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int 创建整数字段
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 创建Int64字段
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Float64 创建浮点数字段
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// Error 创建错误字段
func Error(err error) Field {
	return Field{Key: "error", Value: err.Error()}
}

// Duration 创建持续时间字段
func Duration(key string, duration time.Duration) Field {
	return Field{Key: key, Value: duration.String()}
}

// Logger 日志器
type Logger struct {
	logger      log.Logger
	serviceName string
	level       LogLevel
	baseAttrs   []log.KeyValue
}

// NewLogger 创建新的日志器
func NewLogger(serviceName string, level string) *Logger {
	loggerProvider := global.GetLoggerProvider()
	otelLogger := loggerProvider.Logger(serviceName)

	logLevel := parseLogLevel(level)

	// 获取主机信息
	hostname := getMachineIdentifier()
	hostAddress := getHostAddress()

	// 预创建基础属性
	baseAttrs := []log.KeyValue{
		log.String("service", serviceName),
		log.String("host_id", hostname),
		log.String("host_address", hostAddress),
	}

	return &Logger{
		logger:      otelLogger,
		serviceName: serviceName,
		level:       logLevel,
		baseAttrs:   baseAttrs,
	}
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level string) {
	l.level = parseLogLevel(level)
}

// Debug 调试日志
func (l *Logger) Debug(ctx context.Context, msg string, fields ...Field) {
	if l.level > LevelDebug {
		return
	}
	l.emit(ctx, log.SeverityDebug, msg, fields...)
}

// Info 信息日志
func (l *Logger) Info(ctx context.Context, msg string, fields ...Field) {
	if l.level > LevelInfo {
		return
	}
	l.emit(ctx, log.SeverityInfo, msg, fields...)
}

// Warn 警告日志
func (l *Logger) Warn(ctx context.Context, msg string, fields ...Field) {
	if l.level > LevelWarn {
		return
	}
	l.emit(ctx, log.SeverityWarn, msg, fields...)
}

// Error 错误日志
func (l *Logger) Error(ctx context.Context, msg string, fields ...Field) {
	if l.level > LevelError {
		return
	}
	l.emit(ctx, log.SeverityError, msg, fields...)
}

// ErrorWithErr 记录错误，包含错误对象
func (l *Logger) ErrorWithErr(ctx context.Context, err error, msg string, fields ...Field) {
	if err == nil || l.level > LevelError {
		return
	}

	// 添加错误字段
	allFields := append(fields, Error(err))
	l.emit(ctx, log.SeverityError, msg, allFields...)
}

// emit 发送日志到 OTEL
func (l *Logger) emit(ctx context.Context, severity log.Severity, msg string, fields ...Field) {
	// 复用基础属性，避免重复分配
	attrs := make([]log.KeyValue, 0, len(l.baseAttrs)+len(fields)+3)
	attrs = append(attrs, l.baseAttrs...)
	attrs = append(attrs, log.String("message", msg))

	// 处理额外字段
	for _, field := range fields {
		attrs = append(attrs, log.String(field.Key, fmt.Sprintf("%v", field.Value)))
	}

	// 添加追踪信息（如果存在）
	if span := trace.SpanFromContext(ctx); span.IsRecording() {
		spanCtx := span.SpanContext()
		attrs = append(attrs,
			log.String("trace_id", spanCtx.TraceID().String()),
			log.String("span_id", spanCtx.SpanID().String()),
		)
	}

	// 创建并发送日志记录
	logRecord := log.Record{}
	logRecord.SetTimestamp(time.Now())
	logRecord.SetSeverity(severity)
	logRecord.SetBody(log.StringValue(msg))
	logRecord.AddAttributes(attrs...)

	l.logger.Emit(ctx, logRecord)
}

// getHostAddress 获取主机地址
func getHostAddress() string {
	// 尝试获取本机IP地址
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "unknown"
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return "unknown"
}

// getMachineIdentifier 获取机器唯一标识符
func getMachineIdentifier() string {
	// 1. 首先尝试获取容器ID（Docker环境）
	if containerID := getContainerID(); containerID != "" {
		return containerID[:12] // 使用前12位，类似Docker显示
	}

	// 2. 尝试获取系统machine-id
	if machineID := getMachineID(); machineID != "" {
		return machineID[:8] // 使用前8位作为标识
	}

	// 3. 使用hostname作为备用方案
	if hostname, err := os.Hostname(); err == nil {
		return hostname
	}

	// 4. 最后的备用方案
	return "unknown-host"
}

// getContainerID 获取Docker容器ID
func getContainerID() string {
	// 在Docker容器中，可以从/proc/1/cgroup文件中获取容器ID
	file, err := os.Open("/proc/1/cgroup")
	if err != nil {
		return ""
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.Contains(line, "docker") {
			// 格式通常是: 1:name=systemd:/docker/容器ID
			parts := strings.Split(line, "/")
			if len(parts) > 0 {
				containerID := parts[len(parts)-1]
				if len(containerID) >= 12 {
					return containerID
				}
			}
		}
	}
	return ""
}

// getMachineID 获取系统machine-id
func getMachineID() string {
	// 尝试读取/etc/machine-id
	if content, err := os.ReadFile("/etc/machine-id"); err == nil {
		return strings.TrimSpace(string(content))
	}

	// 尝试读取/var/lib/dbus/machine-id (备用位置)
	if content, err := os.ReadFile("/var/lib/dbus/machine-id"); err == nil {
		return strings.TrimSpace(string(content))
	}

	return ""
}

// parseLogLevel 解析日志级别字符串
func parseLogLevel(level string) LogLevel {
	switch level {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}
