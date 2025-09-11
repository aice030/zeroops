package model

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

// DeployState 发布状态枚举
type DeployState string

const (
	StatusUnrelease DeployState = "unrelease" // 未发布/待发布
	StatusDeploying DeployState = "deploying" // 正在发布
	StatusStop      DeployState = "stop"      // 暂停发布
	StatusRollback  DeployState = "rollback"  // 已回滚
	StatusCompleted DeployState = "completed" // 发布完成
)
