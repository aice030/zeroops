package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"

	"github.com/qiniu/zeroops/internal/service_manager/model"
)

// CreateDeployment 创建发布任务
func (d *Database) CreateDeployment(ctx context.Context, req *model.CreateDeploymentRequest) (string, error) {
	tx, err := d.BeginTx(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	query := `INSERT INTO service_deploy_tasks (service, version, task_creator, deploy_begin_time, deploy_state, correlation_id) 
	          VALUES (?, ?, ?, ?, ?, ?)`

	// 根据是否有计划时间决定初始状态
	var initialStatus model.DeployState
	if req.ScheduleTime == nil {
		initialStatus = model.StatusDeploying // 立即发布
	} else {
		initialStatus = model.StatusUnrelease // 计划发布
	}

	result, err := tx.ExecContext(ctx, query, req.Service, req.Version, "system", req.ScheduleTime, initialStatus, "")
	if err != nil {
		return "", err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return strconv.FormatInt(id, 10), nil
}

// GetDeploymentByID 根据ID获取发布任务详情
func (d *Database) GetDeploymentByID(ctx context.Context, deployID string) (*model.Deployment, error) {
	query := `SELECT id, service, version, deploy_state, deploy_begin_time, deploy_end_time 
	          FROM service_deploy_tasks WHERE id = ?`
	row := d.QueryRowContext(ctx, query, deployID)

	var task model.ServiceDeployTask
	if err := row.Scan(&task.ID, &task.Service, &task.Version, &task.DeployState,
		&task.DeployBeginTime, &task.DeployEndTime); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	deployment := &model.Deployment{
		ID:           strconv.Itoa(int(task.ID)),
		Service:      task.Service,
		Version:      task.Version,
		Status:       task.DeployState,
		ScheduleTime: task.DeployBeginTime,
		FinishTime:   task.DeployEndTime,
	}

	return deployment, nil
}

// GetDeployments 获取发布任务列表
func (d *Database) GetDeployments(ctx context.Context, query *model.DeploymentQuery) ([]model.Deployment, error) {
	sql := `SELECT id, service, version, deploy_state, deploy_begin_time, deploy_end_time 
	        FROM service_deploy_tasks WHERE 1=1`
	args := []any{}

	if query.Type != "" {
		sql += " AND deploy_state = ?"
		args = append(args, query.Type)
	}

	if query.Service != "" {
		sql += " AND service = ?"
		args = append(args, query.Service)
	}

	sql += " ORDER BY deploy_begin_time DESC"

	if query.Limit > 0 {
		sql += " LIMIT ?"
		args = append(args, query.Limit)
	}

	rows, err := d.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []model.Deployment
	for rows.Next() {
		var task model.ServiceDeployTask
		if err := rows.Scan(&task.ID, &task.Service, &task.Version, &task.DeployState,
			&task.DeployBeginTime, &task.DeployEndTime); err != nil {
			return nil, err
		}

		deployment := model.Deployment{
			ID:           strconv.Itoa(int(task.ID)),
			Service:      task.Service,
			Version:      task.Version,
			Status:       task.DeployState,
			ScheduleTime: task.DeployBeginTime,
			FinishTime:   task.DeployEndTime,
		}

		deployments = append(deployments, deployment)
	}

	return deployments, rows.Err()
}

// UpdateDeployment 修改未开始的发布任务
func (d *Database) UpdateDeployment(ctx context.Context, deployID string, req *model.UpdateDeploymentRequest) error {
	sql := `UPDATE service_deploy_tasks SET `
	args := []any{}
	updates := []string{}

	if req.Version != "" {
		updates = append(updates, "version = ?")
		args = append(args, req.Version)
	}

	if req.ScheduleTime != nil {
		updates = append(updates, "deploy_begin_time = ?")
		args = append(args, req.ScheduleTime)
	}

	if len(updates) == 0 {
		return nil
	}

	sql += updates[0]
	for i := 1; i < len(updates); i++ {
		sql += ", " + updates[i]
	}

	sql += " WHERE id = ? AND deploy_state = ?"
	args = append(args, deployID, model.StatusUnrelease)

	_, err := d.ExecContext(ctx, sql, args...)
	return err
}

// DeleteDeployment 删除未开始的发布任务
func (d *Database) DeleteDeployment(ctx context.Context, deployID string) error {
	query := `DELETE FROM service_deploy_tasks WHERE id = ? AND deploy_state = ?`
	_, err := d.ExecContext(ctx, query, deployID, model.StatusUnrelease)
	return err
}

// PauseDeployment 暂停正在灰度的发布任务
func (d *Database) PauseDeployment(ctx context.Context, deployID string) error {
	query := `UPDATE service_deploy_tasks SET deploy_state = ? WHERE id = ? AND deploy_state = ?`
	_, err := d.ExecContext(ctx, query, model.StatusStop, deployID, model.StatusDeploying)
	return err
}

// ContinueDeployment 继续发布
func (d *Database) ContinueDeployment(ctx context.Context, deployID string) error {
	query := `UPDATE service_deploy_tasks SET deploy_state = ? WHERE id = ? AND deploy_state = ?`
	_, err := d.ExecContext(ctx, query, model.StatusDeploying, deployID, model.StatusStop)
	return err
}

// RollbackDeployment 回滚发布任务
func (d *Database) RollbackDeployment(ctx context.Context, deployID string) error {
	query := `UPDATE service_deploy_tasks SET deploy_state = ? WHERE id = ?`
	_, err := d.ExecContext(ctx, query, model.StatusRollback, deployID)
	return err
}

// CheckDeploymentConflict 检查发布冲突
func (d *Database) CheckDeploymentConflict(ctx context.Context, service, version string) (bool, error) {
	query := `SELECT COUNT(*) FROM service_deploy_tasks 
	          WHERE service = ? AND version = ? AND deploy_state IN (?, ?)`
	row := d.QueryRowContext(ctx, query, service, version, model.StatusDeploying, model.StatusStop)

	var count int
	if err := row.Scan(&count); err != nil {
		return false, err
	}

	return count > 0, nil
}

// ===== 部署批次操作 =====

// CreateDeployBatch 创建部署批次
func (d *Database) CreateDeployBatch(ctx context.Context, batch *model.DeployBatch) error {
	nodeIDsJSON, err := json.Marshal(batch.NodeIDs)
	if err != nil {
		return err
	}

	query := `INSERT INTO deploy_batches (deploy_id, batch_id, start_time, end_time, target_ratio, node_ids) 
	          VALUES (?, ?, ?, ?, ?, ?)`
	_, err = d.ExecContext(ctx, query, batch.DeployID, batch.BatchID, batch.StartTime,
		batch.EndTime, batch.TargetRatio, string(nodeIDsJSON))
	return err
}

// GetDeployBatches 获取部署批次列表
func (d *Database) GetDeployBatches(ctx context.Context, deployID int) ([]model.DeployBatch, error) {
	query := `SELECT id, deploy_id, batch_id, start_time, end_time, target_ratio, node_ids 
	          FROM deploy_batches WHERE deploy_id = ? ORDER BY id`
	rows, err := d.QueryContext(ctx, query, deployID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var batches []model.DeployBatch
	for rows.Next() {
		var batch model.DeployBatch
		var nodeIDsJSON string

		if err := rows.Scan(&batch.ID, &batch.DeployID, &batch.BatchID, &batch.StartTime,
			&batch.EndTime, &batch.TargetRatio, &nodeIDsJSON); err != nil {
			return nil, err
		}

		if nodeIDsJSON != "" {
			if err := json.Unmarshal([]byte(nodeIDsJSON), &batch.NodeIDs); err != nil {
				return nil, err
			}
		}

		batches = append(batches, batch)
	}

	return batches, rows.Err()
}
