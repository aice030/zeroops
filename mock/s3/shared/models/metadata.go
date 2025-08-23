package models

import (
	"time"
)

// Metadata 元数据模型
type Metadata struct {
	ID           string            `json:"id" db:"id"`
	Key          string            `json:"key" db:"key"`
	Bucket       string            `json:"bucket" db:"bucket"`
	Size         int64             `json:"size" db:"size"`
	ContentType  string            `json:"content_type" db:"content_type"`
	MD5Hash      string            `json:"md5_hash" db:"md5_hash"`
	ETag         string            `json:"etag" db:"etag"`
	StorageNodes []string          `json:"storage_nodes" db:"storage_nodes"` // JSON 存储
	Headers      map[string]string `json:"headers" db:"headers"`             // JSON 存储
	Tags         map[string]string `json:"tags" db:"tags"`                   // JSON 存储
	Status       string            `json:"status" db:"status"`               // active, deleted, corrupted
	Version      int64             `json:"version" db:"version"`
	CreatedAt    time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time        `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Stats 统计信息
type Stats struct {
	TotalObjects int64            `json:"total_objects"`
	TotalSize    int64            `json:"total_size"`
	AverageSize  float64          `json:"average_size"`
	BucketStats  map[string]int64 `json:"bucket_stats"`
	ContentTypes map[string]int64 `json:"content_types"`
	LastUpdated  time.Time        `json:"last_updated"`
}
