package metrics

import (
	"context"
)

// MetricsCollector 定义指标收集器接口
type MetricsCollector interface {
	// RecordCounter 计数型数据，如请求数、错误数
	RecordCounter(ctx context.Context, name string, value float64, labels map[string]string) error

	// RecordGauge 仪表盘型数据，记录某一时刻的数值，适合用来表示瞬时状态，如当前内存使用、CPU温度等
	RecordGauge(ctx context.Context, name string, value float64, labels map[string]string) error

	// RecordHistogram 直方图类型指标，适合统计一组数据的分布情况，比如请求延迟分布、响应大小分布等。
	RecordHistogram(ctx context.Context, name string, value float64, labels map[string]string) error

	// Reset 清空所有指标
	Reset() error
}

// MetricExporter 负责将采集到的指标数据暴露给 Prometheus 等监控系统
type MetricExporter interface {
	// Start 启动指标暴露服务（例如启动 HTTP 端口暴露 /metrics）
	Start(addr string) error

	// Stop 停止当前指标暴露服务，释放相关资源
	Stop() error
}
