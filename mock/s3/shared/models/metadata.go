package models

import (
	"time"
)

// Metadata 元数据模型
type Metadata struct {
	Bucket      string    `json:"bucket" db:"bucket"`
	Key         string    `json:"key" db:"key"`                   // 对象键
	Size        int64     `json:"size" db:"size"`                 // 文件大小
	ContentType string    `json:"content_type" db:"content_type"` // MIME类型
	MD5Hash     string    `json:"md5_hash" db:"md5_hash"`         // 文件校验和ETag
	Status      string    `json:"status" db:"status"`             // active/deleted/corrupted
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Stats 统计信息
type Stats struct {
	TotalObjects int64     `json:"total_objects"`
	TotalSize    int64     `json:"total_size"`
	LastUpdated  time.Time `json:"last_updated"`
}

// 状态常量
const (
	StatusActive    = "active"
	StatusDeleted   = "deleted"
	StatusCorrupted = "corrupted"
)

// GetID 生成唯一标识符（bucket/key）
func (m *Metadata) GetID() string {
	return m.Bucket + "/" + m.Key
}

// GetETag 获取ETag（使用MD5Hash）
func (m *Metadata) GetETag() string {
	return m.MD5Hash
}
