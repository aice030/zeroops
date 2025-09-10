package deploy_task

import "time"

// DeployState 发布状态枚举
type DeployState string

const (
	StatusDeploying DeployState = "deploying"
	StatusStop      DeployState = "stop"
	StatusRollback  DeployState = "rollback"
	StatusCompleted DeployState = "completed"
)

// DeployBatch 部署批次信息
type DeployBatch struct {
	ID          int        `json:"id" db:"id"`                    // bigint - 主键
	DeployID    int        `json:"deployId" db:"deploy_id"`       // bigint - 外键
	BatchID     string     `json:"batchId" db:"batch_id"`         // varchar(255) - 批次ID
	StartTime   *time.Time `json:"startTime" db:"start_time"`     // datetime
	EndTime     *time.Time `json:"endTime" db:"end_time"`         // datetime
	TargetRatio float64    `json:"targetRatio" db:"target_ratio"` // double
	NodeIDs     []string   `json:"nodeIds" db:"node_ids"`         // 数组格式的节点ID列表
}

// ServiceDeployTask 服务部署任务信息
type ServiceDeployTask struct {
	ID              int         `json:"id" db:"id"`                             // bigint - 主键
	Service         string      `json:"service" db:"service"`                   // varchar(255) - 外键引用services.name
	Version         string      `json:"version" db:"version"`                   // varchar(255) - 外键引用service_versions.version
	TaskCreator     string      `json:"taskCreator" db:"task_creator"`          // varchar(255)
	DeployBeginTime *time.Time  `json:"deployBeginTime" db:"deploy_begin_time"` // datetime
	DeployEndTime   *time.Time  `json:"deployEndTime" db:"deploy_end_time"`     // datetime
	DeployState     DeployState `json:"deployState" db:"deploy_state"`          // 部署状态
	CorrelationID   string      `json:"correlationId" db:"correlation_id"`      // varchar(255)
}

// ===== API请求响应结构体 =====

// Deployment API响应用的发布任务
type Deployment struct {
	ID           string      `json:"id"`
	Service      string      `json:"service"`
	Version      string      `json:"version"`
	Status       DeployState `json:"status"`
	ScheduleTime *time.Time  `json:"scheduleTime,omitempty"`
	FinishTime   *time.Time  `json:"finishTime,omitempty"`
}

// CreateDeploymentRequest 创建发布任务请求
type CreateDeploymentRequest struct {
	Service      string     `json:"service" binding:"required"`
	Version      string     `json:"version" binding:"required"`
	ScheduleTime *time.Time `json:"scheduleTime,omitempty"` // 可选参数，不填为立即发布
}

// UpdateDeploymentRequest 修改发布任务请求
type UpdateDeploymentRequest struct {
	Version      string     `json:"version,omitempty"`
	ScheduleTime *time.Time `json:"scheduleTime,omitempty"` // 新的计划发布时间
}

// DeploymentQuery 发布任务查询参数
type DeploymentQuery struct {
	Type    DeployState `form:"type"`    // deploying/stop/rollback/completed
	Service string      `form:"service"` // 服务名称过滤
	Start   string      `form:"start"`   // 分页起始
	Limit   int         `form:"limit"`   // 分页大小
}
