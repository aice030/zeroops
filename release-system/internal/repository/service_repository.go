package repository

import (
	"context"
	"release-system/internal/model/service_info"
)

// ServiceRepository 服务信息仓储接口
type ServiceRepository interface {
	// GetServicesForAPI 获取所有服务列表（GET /v1/services）
	GetServicesForAPI(ctx context.Context) (*service_info.ServicesResponse, error)

	// GetServiceActiveVersionsForAPI 获取服务活跃版本（GET /v1/services/:service/activeVersions）
	GetServiceActiveVersionsForAPI(ctx context.Context, serviceName string) ([]service_info.ActiveVersionItem, error)

	// GetServiceAvailableVersions 获取可用服务版本（GET /v1/services/:service/availableVersions?type=unrelease）
	// 直接返回ServiceVersion列表，API层负责包装响应格式
	GetServiceAvailableVersions(ctx context.Context, serviceName string, versionType string) ([]service_info.ServiceVersion, error)

	// GetServiceMetrics 获取服务指标数据（GET /v1/services/:service/metricStats）
	GetServiceMetrics(ctx context.Context, serviceName string) (*service_info.MetricStats, error)
}
