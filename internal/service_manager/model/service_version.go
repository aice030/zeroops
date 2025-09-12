package model

import "time"

// ServiceVersion 服务版本信息
type ServiceVersion struct {
	Version    string    `json:"version" db:"version"`        // varchar(255) - 主键
	Service    string    `json:"service" db:"service"`        // varchar(255) - 外键引用services.name
	CreateTime time.Time `json:"createTime" db:"create_time"` // 时间戳字段
}
