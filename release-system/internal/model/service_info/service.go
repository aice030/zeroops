package service_info

import "time"

// ExceptionStatus 异常处理状态枚举
type ExceptionStatus string

const (
	ExceptionStatusNew        ExceptionStatus = "new"
	ExceptionStatusAnalyzing  ExceptionStatus = "analyzing"
	ExceptionStatusProcessing ExceptionStatus = "processing"
	ExceptionStatusResolved   ExceptionStatus = "resolved"
)

// HealthStatus 健康状态枚举
type HealthStatus string

const (
	HealthStatusNormal  HealthStatus = "Normal"
	HealthStatusWarning HealthStatus = "Warning"
	HealthStatusError   HealthStatus = "Error"
)

// DeployStatus 部署状态枚举
type DeployStatus string

const (
	DeployStatusInDeploying     DeployStatus = "InDeploying"
	DeployStatusAllDeployFinish DeployStatus = "AllDeployFinish"
)

// Service 服务基础信息
type Service struct {
	Name string   `json:"name" db:"name"` // varchar(255) - 主键
	Deps []string `json:"deps" db:"deps"` // 依赖关系
}

// ServiceInstance 服务实例信息
type ServiceInstance struct {
	ID      string `json:"id" db:"id"`           // 主键
	Service string `json:"service" db:"service"` // varchar(255) - 外键引用services.name
	Version string `json:"version" db:"version"` // varchar(255) - 外键引用service_versions.version
}

// ServiceVersion 服务版本信息
type ServiceVersion struct {
	Version    string    `json:"version" db:"version"`        // varchar(255) - 主键
	Service    string    `json:"service" db:"service"`        // varchar(255) - 外键引用services.name
	CreateTime time.Time `json:"createTime" db:"create_time"` // 时间戳字段
}

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

// ===== API响应结构体 =====

// ServiceItem API响应用的服务信息（对应/v1/services接口items格式）
type ServiceItem struct {
	Name        string       `json:"name"`        // 服务名称
	DeployState DeployStatus `json:"deployState"` // 发布状态：InDeploying|AllDeployFinish
	Health      HealthStatus `json:"health"`      // 健康状态：Normal/Warning/Error
	Deps        []string     `json:"deps"`        // 依赖关系（直接使用Service.Deps）
}

// ServicesResponse 服务列表API响应（对应/v1/services接口）
type ServicesResponse struct {
	Items    []ServiceItem       `json:"items"`
	Relation map[string][]string `json:"relation"` // 树形关系描述，有向无环图
}

// ActiveVersionItem 活跃版本项目
type ActiveVersionItem struct {
	Version                 string       `json:"version"`                 // v1.0.1
	DeployID                string       `json:"deployID"`                // 1001
	StartTime               time.Time    `json:"startTime"`               // 开始时间
	EstimatedCompletionTime time.Time    `json:"estimatedCompletionTime"` // 预估完成时间
	Instances               int          `json:"instances"`               // 实例个数
	Health                  HealthStatus `json:"health"`                  // 健康状态：Normal/Warning/Error
}

// MetricStats 服务指标统计（对应/v1/services/:service/metricStats接口）
type MetricStats struct {
	Summary MetricSummary       `json:"summary"` // 所有实例的聚合值
	Items   []MetricVersionItem `json:"items"`   // 各版本的指标内容
}

// MetricSummary 指标汇总
type MetricSummary struct {
	Metrics []Metric `json:"metrics"` // 此版本发布的metric指标内容
}

// MetricVersionItem 版本指标项
type MetricVersionItem struct {
	Version string   `json:"version"` // v1.0.1
	Metrics []Metric `json:"metrics"` // 此版本发布的metric指标内容
}

// Metric 指标
type Metric struct {
	Name  string `json:"name"`  // latency/traffic/errorRatio/saturation
	Value any    `json:"value"` // 指标值（ms/Qps/百分比）
}
