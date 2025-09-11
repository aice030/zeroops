package service

import (
	"context"

	"github.com/qiniu/zeroops/internal/service_manager/model"
	"github.com/rs/zerolog/log"
)

// ===== 部署管理业务方法 =====

// CreateDeployment 创建发布任务
func (s *Service) CreateDeployment(ctx context.Context, req *model.CreateDeploymentRequest) (string, error) {
	// 检查服务是否存在
	service, err := s.db.GetServiceByName(ctx, req.Service)
	if err != nil {
		return "", err
	}
	if service == nil {
		return "", ErrServiceNotFound
	}

	// 检查发布冲突
	conflict, err := s.db.CheckDeploymentConflict(ctx, req.Service, req.Version)
	if err != nil {
		return "", err
	}
	if conflict {
		return "", ErrDeploymentConflict
	}

	// 创建发布任务
	deployID, err := s.db.CreateDeployment(ctx, req)
	if err != nil {
		return "", err
	}

	log.Info().
		Str("deployID", deployID).
		Str("service", req.Service).
		Str("version", req.Version).
		Msg("deployment created successfully")

	return deployID, nil
}

// GetDeploymentByID 获取发布任务详情
func (s *Service) GetDeploymentByID(ctx context.Context, deployID string) (*model.Deployment, error) {
	deployment, err := s.db.GetDeploymentByID(ctx, deployID)
	if err != nil {
		return nil, err
	}
	if deployment == nil {
		return nil, ErrDeploymentNotFound
	}
	return deployment, nil
}

// GetDeployments 获取发布任务列表
func (s *Service) GetDeployments(ctx context.Context, query *model.DeploymentQuery) ([]model.Deployment, error) {
	return s.db.GetDeployments(ctx, query)
}

// UpdateDeployment 修改发布任务
func (s *Service) UpdateDeployment(ctx context.Context, deployID string, req *model.UpdateDeploymentRequest) error {
	// 检查部署任务是否存在
	deployment, err := s.db.GetDeploymentByID(ctx, deployID)
	if err != nil {
		return err
	}
	if deployment == nil {
		return ErrDeploymentNotFound
	}

	// 只有unrelease状态的任务可以修改
	if deployment.Status != model.StatusUnrelease {
		return ErrInvalidDeployState
	}

	err = s.db.UpdateDeployment(ctx, deployID, req)
	if err != nil {
		return err
	}

	log.Info().
		Str("deployID", deployID).
		Str("service", deployment.Service).
		Msg("deployment updated successfully")

	return nil
}

// DeleteDeployment 删除发布任务
func (s *Service) DeleteDeployment(ctx context.Context, deployID string) error {
	// 检查部署任务是否存在
	deployment, err := s.db.GetDeploymentByID(ctx, deployID)
	if err != nil {
		return err
	}
	if deployment == nil {
		return ErrDeploymentNotFound
	}

	// 只有未开始的任务可以删除
	if deployment.Status != model.StatusUnrelease {
		return ErrInvalidDeployState
	}

	err = s.db.DeleteDeployment(ctx, deployID)
	if err != nil {
		return err
	}

	log.Info().
		Str("deployID", deployID).
		Str("service", deployment.Service).
		Msg("deployment deleted successfully")

	return nil
}

// PauseDeployment 暂停发布任务
func (s *Service) PauseDeployment(ctx context.Context, deployID string) error {
	// 检查部署任务是否存在且为正在部署状态
	deployment, err := s.db.GetDeploymentByID(ctx, deployID)
	if err != nil {
		return err
	}
	if deployment == nil {
		return ErrDeploymentNotFound
	}
	if deployment.Status != model.StatusDeploying {
		return ErrInvalidDeployState
	}

	err = s.db.PauseDeployment(ctx, deployID)
	if err != nil {
		return err
	}

	log.Info().
		Str("deployID", deployID).
		Str("service", deployment.Service).
		Msg("deployment paused successfully")

	return nil
}

// ContinueDeployment 继续发布任务
func (s *Service) ContinueDeployment(ctx context.Context, deployID string) error {
	// 检查部署任务是否存在且为暂停状态
	deployment, err := s.db.GetDeploymentByID(ctx, deployID)
	if err != nil {
		return err
	}
	if deployment == nil {
		return ErrDeploymentNotFound
	}
	if deployment.Status != model.StatusStop {
		return ErrInvalidDeployState
	}

	err = s.db.ContinueDeployment(ctx, deployID)
	if err != nil {
		return err
	}

	log.Info().
		Str("deployID", deployID).
		Str("service", deployment.Service).
		Msg("deployment continued successfully")

	return nil
}

// RollbackDeployment 回滚发布任务
func (s *Service) RollbackDeployment(ctx context.Context, deployID string) error {
	// 检查部署任务是否存在
	deployment, err := s.db.GetDeploymentByID(ctx, deployID)
	if err != nil {
		return err
	}
	if deployment == nil {
		return ErrDeploymentNotFound
	}

	// 只有正在部署或暂停的任务可以回滚
	if deployment.Status != model.StatusDeploying && deployment.Status != model.StatusStop {
		return ErrInvalidDeployState
	}

	err = s.db.RollbackDeployment(ctx, deployID)
	if err != nil {
		return err
	}

	log.Info().
		Str("deployID", deployID).
		Str("service", deployment.Service).
		Msg("deployment rolled back successfully")

	return nil
}

// GetDeployBatches 获取部署批次列表
func (s *Service) GetDeployBatches(ctx context.Context, deployID string) ([]model.DeployBatch, error) {
	// 先验证部署任务存在
	deployment, err := s.db.GetDeploymentByID(ctx, deployID)
	if err != nil {
		return nil, err
	}
	if deployment == nil {
		return nil, ErrDeploymentNotFound
	}

	// TODO:将deployID转换为int (这里简化处理)
	// 实际项目中应该在数据库层统一ID类型
	return s.db.GetDeployBatches(ctx, 1) // 临时硬编码，需要根据实际deployID转换
}

// CreateDeployBatch 创建部署批次
func (s *Service) CreateDeployBatch(ctx context.Context, batch *model.DeployBatch) error {
	return s.db.CreateDeployBatch(ctx, batch)
}
