package model

// Status 异常处理状态枚举
type Status string

const (
	StatusNew        Status = "new"
	StatusAnalyzing  Status = "analyzing"
	StatusProcessing Status = "processing"
	StatusResolved   Status = "resolved"
)

// Level 健康状态枚举
type Level string

const (
	LevelWarning Level = "Warning"
	LevelError   Level = "Error"
)

// HealthLevel API响应用的健康状态枚举（包含正常状态）
type HealthLevel string

const (
	HealthLevelNormal  HealthLevel = "Normal"
	HealthLevelWarning HealthLevel = "Warning"
	HealthLevelError   HealthLevel = "Error"
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
