package adapter

import "time"

// ExternalDeploySystem 公司发布系统接口定义
type ExternalDeploySystem interface {
	// CreateDeployment 调用公司发布系统创建发布任务
	CreateDeployment(req *ExternalDeployRequest) (*ExternalDeployResponse, error)

	// PauseDeployment 暂停发布任务
	PauseDeployment(deployID string) error

	// ContinueDeployment 继续发布任务
	ContinueDeployment(deployID string) error

	// RollbackDeployment 回滚发布任务
	RollbackDeployment(deployID string) error

	// GetDeploymentStatus 获取发布状态
	GetDeploymentStatus(deployID string) (*ExternalDeployStatus, error)
}

// ExternalAlertSystem 公司告警系统接口定义
type ExternalAlertSystem interface {
	// SendAlert 发送告警
	SendAlert(alert *AlertRequest) error

	// GetAlerts 获取告警列表
	GetAlerts(query *AlertQuery) (*AlertResponse, error)
}

// ExternalDeployRequest 公司发布系统请求
type ExternalDeployRequest struct {
	Service      string         `json:"service"`
	Version      string         `json:"version"`
	ScheduleTime *time.Time     `json:"scheduleTime,omitempty"`
	Config       map[string]any `json:"config,omitempty"`
}

// ExternalDeployResponse 公司发布系统响应
type ExternalDeployResponse struct {
	DeployID string `json:"deployId"`
	Status   string `json:"status"`
	Message  string `json:"message,omitempty"`
}

// ExternalDeployStatus 公司发布状态
type ExternalDeployStatus struct {
	DeployID     string     `json:"deployId"`
	Service      string     `json:"service"`
	Version      string     `json:"version"`
	Status       string     `json:"status"`
	Progress     float64    `json:"progress"` // 发布进度 0-100
	StartTime    *time.Time `json:"startTime,omitempty"`
	FinishTime   *time.Time `json:"finishTime,omitempty"`
	ErrorMessage string     `json:"errorMessage,omitempty"`
}

// AlertRequest 告警请求
type AlertRequest struct {
	Title       string            `json:"title"`
	Content     string            `json:"content"`
	Level       string            `json:"level"` // Info/Warning/Error/Critical
	Service     string            `json:"service"`
	DeployID    string            `json:"deployId,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]any    `json:"annotations,omitempty"`
}

// AlertQuery 告警查询参数
type AlertQuery struct {
	Service string     `json:"service,omitempty"`
	Level   string     `json:"level,omitempty"`
	Start   *time.Time `json:"start,omitempty"`
	End     *time.Time `json:"end,omitempty"`
	Limit   int        `json:"limit,omitempty"`
}

// AlertResponse 告警响应
type AlertResponse struct {
	Items []Alert `json:"items"`
	Total int     `json:"total"`
}

// Alert 告警信息
type Alert struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Content     string            `json:"content"`
	Level       string            `json:"level"`
	Service     string            `json:"service"`
	Status      string            `json:"status"` // Active/Resolved
	CreatedAt   time.Time         `json:"createdAt"`
	ResolvedAt  *time.Time        `json:"resolvedAt,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]any    `json:"annotations,omitempty"`
}

// ===== 数据库模型（对应ER图中的表） =====

// EventLog 事件日志数据库模型（对应event_logs表）
type EventLog struct {
	ID            int       `json:"id" db:"id"`                        // bigint - 主键
	SourceSystem  string    `json:"sourceSystem" db:"source_system"`   // varchar(255)
	SourceID      string    `json:"sourceId" db:"source_id"`           // varchar(255)
	SourceName    string    `json:"sourceName" db:"source_name"`       // varchar(255)
	Timestamp     time.Time `json:"timestamp" db:"timestamp"`          // datetime
	Actor         string    `json:"actor" db:"actor"`                  // varchar(255)
	Severity      string    `json:"severity" db:"severity"`            // varchar(255)
	CorrelationID string    `json:"correlationId" db:"correlation_id"` // varchar(255)
	Payload       string    `json:"payload" db:"payload"`              // text - JSON格式
}

// MetricAlertChange 指标告警变更数据库模型（对应metric_alert_changes表）
type MetricAlertChange struct {
	ID         int       `json:"id" db:"id"`                  // bigint - 主键
	ChangeTime time.Time `json:"changeTime" db:"change_time"` // datetime
	AlertName  string    `json:"alertName" db:"alert_name"`   // varchar(255)
	ChangeItem string    `json:"changeItem" db:"change_item"` // text - 变更项内容
}
