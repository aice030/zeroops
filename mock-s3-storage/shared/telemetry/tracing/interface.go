package tracing

import (
	"context"
	"time"
)

// Span 代表一个链路跨度，封装底层 tracing 细节
type Span interface {
	End(options ...SpanEndOption)
	AddEvent(name string, attributes map[string]interface{})
	SetAttribute(key string, value interface{})
	SetStatus(code StatusCode, description string)
	Context() context.Context
}

// StatusCode 定义 Span 状态码
type StatusCode int

const (
	StatusCodeUnset StatusCode = iota
	StatusCodeOk
	StatusCodeError
)

// SpanEndOption 结束 Span 的可选参数
type SpanEndOption func(*spanEndConfig)

type spanEndConfig struct {
	endTime time.Time
}

// WithEndTime 指定 Span 结束时间
func WithEndTime(t time.Time) SpanEndOption {
	return func(c *spanEndConfig) {
		c.endTime = t
	}
}

// Tracer 统一链路跟踪接口
type Tracer interface {
	// StartSpan 开启一个新的 Span，返回 Span 对象和新 Context
	StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span)

	// Inject 用于跨进程传播 TraceContext（HTTP Header等）
	Inject(ctx context.Context, carrier interface{}) error

	// Extract 从跨进程载体中提取 TraceContext
	Extract(ctx context.Context, carrier interface{}) (context.Context, error)
}

// SpanOption Span 选项占位符（可扩展）
type SpanOption func(*spanOptions)

type spanOptions struct {
	// 预留配置，如 Span 类型、采样率等
}
