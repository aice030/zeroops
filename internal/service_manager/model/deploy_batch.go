package model

import "time"

// DeployBatch 部署批次信息
type DeployBatch struct {
	ID          int64      `json:"id" db:"id"`                    // bigint - 主键
	DeployID    int64      `json:"deployId" db:"deploy_id"`       // bigint - 外键
	BatchID     string     `json:"batchId" db:"batch_id"`         // varchar(255) - 批次ID
	StartTime   *time.Time `json:"startTime" db:"start_time"`     // datetime
	EndTime     *time.Time `json:"endTime" db:"end_time"`         // datetime
	TargetRatio float64    `json:"targetRatio" db:"target_ratio"` // double
	NodeIDs     []string   `json:"nodeIds" db:"node_ids"`         // 数组格式的节点ID列表
}
