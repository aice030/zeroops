package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"time"

	"github.com/qiniu/zeroops/internal/service_manager/model"
)

// CreateDeployment 创建发布任务
func (d *Database) CreateDeployment(ctx context.Context, req *model.CreateDeploymentRequest) (string, error) {
	// 生成唯一ID
	deployID := "deploy-" + strconv.FormatInt(time.Now().UnixNano(), 36)

	// 根据是否有计划时间决定初始状态
	var initialStatus model.DeployState
	if req.ScheduleTime == nil {
		initialStatus = model.StatusDeploying // 立即发布
	} else {
		initialStatus = model.StatusUnrelease // 计划发布
	}

	query := `INSERT INTO deploy_tasks (id, start_time, end_time, target_ratio, instances, deploy_state) 
	          VALUES ($1, $2, $3, $4, $5, $6)`

	// 默认实例为空数组
	instances := []string{}
	instancesJSON, _ := json.Marshal(instances)

	_, err := d.ExecContext(ctx, query, deployID, req.ScheduleTime, nil, 0.0, string(instancesJSON), initialStatus)
	if err != nil {
		return "", err
	}

	return deployID, nil
}

// GetDeploymentByID 根据ID获取发布任务详情
func (d *Database) GetDeploymentByID(ctx context.Context, deployID string) (*model.Deployment, error) {
	query := `SELECT id, start_time, end_time, target_ratio, instances, deploy_state 
	          FROM deploy_tasks WHERE id = $1`
	row := d.QueryRowContext(ctx, query, deployID)

	var task model.ServiceDeployTask
	var instancesJSON string
	if err := row.Scan(&task.ID, &task.StartTime, &task.EndTime, &task.TargetRatio,
		&instancesJSON, &task.DeployState); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// 解析实例JSON数组
	if instancesJSON != "" {
		if err := json.Unmarshal([]byte(instancesJSON), &task.Instances); err != nil {
			return nil, err
		}
	}

	deployment := &model.Deployment{
		ID:           task.ID,
		Status:       task.DeployState,
		ScheduleTime: task.StartTime,
		FinishTime:   task.EndTime,
	}

	return deployment, nil
}

// GetDeployments 获取发布任务列表
func (d *Database) GetDeployments(ctx context.Context, query *model.DeploymentQuery) ([]model.Deployment, error) {
	sql := `SELECT id, start_time, end_time, target_ratio, instances, deploy_state 
	        FROM deploy_tasks WHERE 1=1`
	args := []any{}

	if query.Type != "" {
		sql += " AND deploy_state = $" + strconv.Itoa(len(args)+1)
		args = append(args, query.Type)
	}

	// 注意：新的deploy_tasks表没有service字段，暂时忽略service过滤
	// TODO: 需要根据业务逻辑决定如何处理service过滤

	sql += " ORDER BY start_time DESC"

	if query.Limit > 0 {
		sql += " LIMIT $" + strconv.Itoa(len(args)+1)
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
		var instancesJSON string
		if err := rows.Scan(&task.ID, &task.StartTime, &task.EndTime, &task.TargetRatio,
			&instancesJSON, &task.DeployState); err != nil {
			return nil, err
		}

		// 解析实例JSON数组
		if instancesJSON != "" {
			if err := json.Unmarshal([]byte(instancesJSON), &task.Instances); err != nil {
				return nil, err
			}
		}

		deployment := model.Deployment{
			ID:           task.ID,
			Status:       task.DeployState,
			ScheduleTime: task.StartTime,
			FinishTime:   task.EndTime,
		}

		deployments = append(deployments, deployment)
	}

	return deployments, rows.Err()
}

// UpdateDeployment 修改未开始的发布任务
func (d *Database) UpdateDeployment(ctx context.Context, deployID string, req *model.UpdateDeploymentRequest) error {
	sql := `UPDATE deploy_tasks SET `
	args := []any{}
	updates := []string{}
	paramIndex := 1

	// 注意：新的deploy_tasks表没有version字段，暂时忽略version更新
	// TODO: 需要根据业务逻辑决定如何处理version更新

	if req.ScheduleTime != nil {
		updates = append(updates, "start_time = $"+strconv.Itoa(paramIndex))
		args = append(args, req.ScheduleTime)
		paramIndex++
	}

	if len(updates) == 0 {
		return nil
	}

	sql += updates[0]
	for i := 1; i < len(updates); i++ {
		sql += ", " + updates[i]
	}

	sql += " WHERE id = $" + strconv.Itoa(paramIndex) + " AND deploy_state = $" + strconv.Itoa(paramIndex+1)
	args = append(args, deployID, model.StatusUnrelease)

	_, err := d.ExecContext(ctx, sql, args...)
	return err
}

// DeleteDeployment 删除未开始的发布任务
func (d *Database) DeleteDeployment(ctx context.Context, deployID string) error {
	query := `DELETE FROM deploy_tasks WHERE id = $1 AND deploy_state = $2`
	_, err := d.ExecContext(ctx, query, deployID, model.StatusUnrelease)
	return err
}

// PauseDeployment 暂停正在灰度的发布任务
func (d *Database) PauseDeployment(ctx context.Context, deployID string) error {
	query := `UPDATE deploy_tasks SET deploy_state = $1 WHERE id = $2 AND deploy_state = $3`
	_, err := d.ExecContext(ctx, query, model.StatusStop, deployID, model.StatusDeploying)
	return err
}

// ContinueDeployment 继续发布
func (d *Database) ContinueDeployment(ctx context.Context, deployID string) error {
	query := `UPDATE deploy_tasks SET deploy_state = $1 WHERE id = $2 AND deploy_state = $3`
	_, err := d.ExecContext(ctx, query, model.StatusDeploying, deployID, model.StatusStop)
	return err
}

// RollbackDeployment 回滚发布任务
func (d *Database) RollbackDeployment(ctx context.Context, deployID string) error {
	query := `UPDATE deploy_tasks SET deploy_state = $1 WHERE id = $2`
	_, err := d.ExecContext(ctx, query, model.StatusRollback, deployID)
	return err
}

// CheckDeploymentConflict 检查发布冲突
// 注意：新的deploy_tasks表没有service和version字段，这个方法需要重新设计
// TODO: 需要根据业务逻辑决定如何检查部署冲突
func (d *Database) CheckDeploymentConflict(ctx context.Context, service, version string) (bool, error) {
	// 暂时返回false，表示没有冲突
	// 实际业务中可能需要通过其他方式检查冲突，比如检查正在部署的任务数量
	return false, nil
}
