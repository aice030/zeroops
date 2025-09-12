package model

import "time"

// ServiceDeployTask 服务部署任务信息
type ServiceDeployTask struct {
	ID              int64       `json:"id" db:"id"`                             // bigint - 主键
	Service         string      `json:"service" db:"service"`                   // varchar(255) - 外键引用services.name
	Version         string      `json:"version" db:"version"`                   // varchar(255) - 外键引用service_versions.version
	TaskCreator     string      `json:"taskCreator" db:"task_creator"`          // varchar(255)
	DeployBeginTime *time.Time  `json:"deployBeginTime" db:"deploy_begin_time"` // datetime
	DeployEndTime   *time.Time  `json:"deployEndTime" db:"deploy_end_time"`     // datetime
	DeployState     DeployState `json:"deployState" db:"deploy_state"`          // 部署状态
	CorrelationID   string      `json:"correlationId" db:"correlation_id"`      // varchar(255)
}
