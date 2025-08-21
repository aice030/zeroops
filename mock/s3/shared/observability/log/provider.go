package log

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// LoggerProviderConfig Logger Provider 配置
type LoggerProviderConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	OTLPEndpoint   string
}

// LoggerProvider OTEL Logger Provider
type LoggerProvider struct {
	provider *sdklog.LoggerProvider
}

// InitializeLoggerProvider 初始化 OTEL Logger Provider
func InitializeLoggerProvider(config *LoggerProviderConfig) (*LoggerProvider, error) {
	// 创建资源
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
			semconv.DeploymentEnvironment(config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// 创建 OTLP HTTP 导出器
	exporter, err := otlploghttp.New(context.Background(),
		otlploghttp.WithEndpoint(config.OTLPEndpoint),
		otlploghttp.WithInsecure(), // 开发环境使用HTTP
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP log exporter: %w", err)
	}

	// 创建 Logger Provider
	provider := sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
	)

	// 设置全局 Logger Provider
	global.SetLoggerProvider(provider)

	return &LoggerProvider{
		provider: provider,
	}, nil
}

// GetLoggerProvider 获取 Logger Provider
func (lp *LoggerProvider) GetLoggerProvider() log.LoggerProvider {
	return lp.provider
}

// Shutdown 关闭 Logger Provider
func (lp *LoggerProvider) Shutdown(ctx context.Context) error {
	return lp.provider.Shutdown(ctx)
}

// NewDefaultLoggerProvider 创建默认的 Logger Provider
func NewDefaultLoggerProvider(serviceName string) (*LoggerProvider, error) {
	config := &LoggerProviderConfig{
		ServiceName:    serviceName,
		ServiceVersion: getEnv("SERVICE_VERSION", "1.0.0"),
		Environment:    getEnv("ENVIRONMENT", "development"),
		OTLPEndpoint:   getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4318"),
	}

	return InitializeLoggerProvider(config)
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
