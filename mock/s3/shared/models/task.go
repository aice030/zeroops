package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DeleteTask 删除文件任务
type DeleteTask struct {
	ID        string     `json:"id"`
	ObjectKey string     `json:"object_key"` // 要删除的对象键
	Status    TaskStatus `json:"status"`     // pending/completed/failed
	Error     string     `json:"error,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// SaveTask 保存文件任务（用于第三方对象保存到本地）
type SaveTask struct {
	ID        string     `json:"id"`
	ObjectKey string     `json:"object_key"` // 对象键
	Object    *Object    `json:"object"`     // 要保存的对象数据
	Status    TaskStatus `json:"status"`     // pending/completed/failed
	Error     string     `json:"error,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// GenerateID 生成删除任务ID
func (t *DeleteTask) GenerateID() {
	if t.ID == "" {
		t.ID = fmt.Sprintf("del_%s", uuid.New().String())
	}
}

// GenerateID 生成保存任务ID
func (t *SaveTask) GenerateID() {
	if t.ID == "" {
		t.ID = fmt.Sprintf("save_%s", uuid.New().String())
	}
}

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)
