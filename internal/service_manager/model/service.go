package model

// Service 服务基础信息
type Service struct {
	Name string   `json:"name" db:"name"` // varchar(255) - 主键
	Deps []string `json:"deps" db:"deps"` // 依赖关系
}
