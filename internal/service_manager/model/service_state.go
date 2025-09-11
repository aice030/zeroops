package model

import "time"

// ServiceState 服务状态信息
type ServiceState struct {
	Service         string          `json:"service" db:"service"`                  // varchar(255) - 外键引用services.name
	Version         string          `json:"version" db:"version"`                  // varchar(255) - 外键引用service_versions.version
	Level           string          `json:"level" db:"level"`                      // varchar(255) - 异常级别
	ReportAt        time.Time       `json:"reportAt" db:"report_at"`               // datetime - 报告时间
	ResolvedAt      *time.Time      `json:"resolvedAt" db:"resolved_at"`           // datetime - 解决时间（可为null）
	HealthStatus    HealthStatus    `json:"healthStatus" db:"health_status"`       // varchar(255) - 健康状态
	ExceptionStatus ExceptionStatus `json:"exceptionStatus" db:"exception_status"` // varchar(255) - 异常处理状态
	Details         string          `json:"details" db:"details"`                  // text - JSON格式的详细信息
}
