package model

// ServiceInstance 服务实例信息
type ServiceInstance struct {
	ID      string `json:"id" db:"id"`           // 主键
	Service string `json:"service" db:"service"` // varchar(255) - 外键引用services.name
	Version string `json:"version" db:"version"` // varchar(255) - 外键引用service_versions.version
}
