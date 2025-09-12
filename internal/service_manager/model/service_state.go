package model

import "time"

// ServiceState 服务状态信息
type ServiceState struct {
	Service       string     `json:"service" db:"service"`              // varchar(255) - 联合PK
	Version       string     `json:"version" db:"version"`              // varchar(255) - 联合PK
	Level         Level      `json:"level" db:"level"`                  // 异常级别
	Detail        string     `json:"detail" db:"detail"`                // text - 详细信息
	ReportAt      time.Time  `json:"reportAt" db:"report_at"`           // time - 报告时间
	ResolvedAt    *time.Time `json:"resolvedAt" db:"resolved_at"`       // time - 解决时间
	HealthStatus  Status     `json:"healthStatus" db:"health_status"`   // 健康状态
	CorrelationID string     `json:"correlationId" db:"correlation_id"` // varchar - 关联ID
}
