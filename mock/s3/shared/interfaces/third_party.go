package interfaces

import (
	"context"
	"mocks3/shared/models"
)

// ThirdPartyService 第三方服务接口
type ThirdPartyService interface {
	// 对象操作
	GetObject(ctx context.Context, bucket, key string) (*models.Object, error)

	// 健康检查
	HealthCheck(ctx context.Context) error
}
