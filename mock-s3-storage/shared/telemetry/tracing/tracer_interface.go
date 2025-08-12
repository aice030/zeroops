package tracing

import (
	"context"
	"net/http"
	"shared/config"
)

// Tracer 链路追踪接口
type Tracer interface {
	// 开始span
	StartSpan(ctx context.Context, name string) (context.Context, Span)

	// 从HTTP请求提取trace信息
	Extract(ctx context.Context, headers http.Header) context.Context

	// 注入trace信息到HTTP请求
	Inject(ctx context.Context, headers http.Header) error
}

// Span 链路跨度
type Span interface {
	End()
	SetAttribute(key string, value any)
	SetError(err error)
}

// NewTracer 创建链路追踪器
func NewTracer(config config.TracingConfig) Tracer {
	// TODO: 实现链路追踪器
	return nil
}
