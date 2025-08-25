package interfaces

import (
	"context"
	"mocks3/shared/models"
)

// StorageService 存储服务接口
type StorageService interface {
	// 公共API - 完整业务流程
	WriteObject(ctx context.Context, object *models.Object) error
	ReadObject(ctx context.Context, bucket, key string) (*models.Object, error)
	DeleteObject(ctx context.Context, bucket, key string) error
	ListObjects(ctx context.Context, req *models.ListObjectsRequest) (*models.ListObjectsResponse, error)

	// 内部API - 仅操作存储层（供Queue Service使用）
	WriteObjectToStorage(ctx context.Context, object *models.Object) error
	DeleteObjectFromStorage(ctx context.Context, bucket, key string) error

	// 统计信息
	GetStats(ctx context.Context) (map[string]any, error)

	// 健康检查
	HealthCheck(ctx context.Context) error
}

// StorageNode 存储节点接口
type StorageNode interface {
	GetNodeID() string
	Write(ctx context.Context, object *models.Object) error
	Read(ctx context.Context, bucket, key string) (*models.Object, error)
	Delete(ctx context.Context, bucket, key string) error
	IsHealthy(ctx context.Context) bool
}

// StorageManager 存储管理器接口
type StorageManager interface {
	AddNode(node StorageNode)
	WriteToAllNodes(ctx context.Context, object *models.Object) error
	ReadFromBestNode(ctx context.Context, bucket, key string) (*models.Object, error)
	DeleteFromAllNodes(ctx context.Context, bucket, key string) error
	GetHealthyNodes() []StorageNode
}

// NodeStatus 存储节点状态
type NodeStatus struct {
	ID        string `json:"id"`
	Status    string `json:"status"` // healthy, unhealthy, unreachable, error
	UsedSpace int64  `json:"used_space"`
	Error     string `json:"error,omitempty"`
}
