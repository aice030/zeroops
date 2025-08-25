package models

import (
	"fmt"
	"time"
)

// DeleteTask 删除文件任务
type DeleteTask struct {
	ID        string     `json:"id"`
	ObjectKey string     `json:"object_key"` // 要删除的对象键
	Status    TaskStatus `json:"status"`     // pending/completed/failed
	Error     string     `json:"error,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// GenerateID 生成任务ID
func (t *DeleteTask) GenerateID() {
	if t.ID == "" {
		t.ID = fmt.Sprintf("del_%d", time.Now().UnixNano())
	}
}

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)
