package model

import "time"

// ServiceDeployTask 服务部署任务信息
type ServiceDeployTask struct {
	ID          string      `json:"id" db:"id"`                    // varchar(32) - 主键
	StartTime   *time.Time  `json:"startTime" db:"start_time"`     // time - 开始时间
	EndTime     *time.Time  `json:"endTime" db:"end_time"`         // time - 结束时间
	TargetRatio float64     `json:"targetRatio" db:"target_ratio"` // double(指导值) - 目标比例
	Instances   []string    `json:"instances" db:"instances"`      // array(真实发布的节点列表) - 实例列表
	DeployState DeployState `json:"deployState" db:"deploy_state"` // 部署状态
}
