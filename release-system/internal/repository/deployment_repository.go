package repository

import (
	"context"
	"release-system/internal/model/deploy_task"
)

// DeploymentRepository 发布任务仓储接口
type DeploymentRepository interface {
	// ===== API方法 =====

	// CreateDeployment 创建发布任务（POST /v1/deployments）
	// 直接返回部署任务ID
	CreateDeployment(ctx context.Context, req *deploy_task.CreateDeploymentRequest) (string, error)

	// GetDeploymentByID 根据ID获取发布任务详情（GET /v1/deployments/:deployID）
	GetDeploymentByID(ctx context.Context, deployID string) (*deploy_task.Deployment, error)

	// GetDeployments 获取发布任务列表（GET /v1/deployments?type=Schedule&service=xxx）
	GetDeployments(ctx context.Context, query *deploy_task.DeploymentQuery) ([]deploy_task.Deployment, error)

	// UpdateDeployment 修改未开始的发布任务（POST /v1/deployments/:deployID）
	UpdateDeployment(ctx context.Context, deployID string, req *deploy_task.UpdateDeploymentRequest) error

	// DeleteDeployment 删除未开始的发布任务（DELETE /v1/deployments/:deployID）
	DeleteDeployment(ctx context.Context, deployID string) error

	// PauseDeployment 暂停正在灰度的发布任务（POST /v1/deployments/:deployID/pause）
	PauseDeployment(ctx context.Context, deployID string) error

	// ContinueDeployment 继续发布（POST /v1/deployments/:deployID/continue）
	ContinueDeployment(ctx context.Context, deployID string) error

	// RollbackDeployment 回滚发布任务（POST /v1/deployments/:deployID/rollback）
	RollbackDeployment(ctx context.Context, deployID string) error

	// CheckDeploymentConflict 检查发布冲突（同一服务同一版本是否已在发布）
	CheckDeploymentConflict(ctx context.Context, service, version string) (bool, error)
}
