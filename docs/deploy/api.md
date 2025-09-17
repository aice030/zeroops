# 发布系统接口参考文档

## 基础信息

本文档描述发布系统的本地函数调用接口，可直接在Go代码中调用。

## 1. 发布相关接口

### 1.1 触发发布

**函数描述**: 触发指定服务版本的发布操作

**函数签名**:
```go
func ExecuteDeployment(params *DeployParams) (*DeployResult, error)
```

**输入参数**:
```go
type DeployParams struct {
    Service    string   `json:"service"`     // 必填，服务名称
    Version    string   `json:"version"`     // 必填，目标版本号
    Instances  []string `json:"instances"`   // 必填，实例ID列表
    PackageURL string   `json:"package_url"` // 必填，包下载URL
    DeployID   string   `json:"deploy_id"`   // 必填，发布任务ID
    Timeout    int      `json:"timeout"`     // 可选，超时时间（秒），默认300
    RetryCount int      `json:"retry_count"` // 可选，重试次数，默认3
}
```

**参数说明**:
- `Service`: 服务名称，如 "user-service"
- `Version`: 版本号，如 "v1.2.3"
- `Instances`: 实例ID数组，如 ["instance-1", "instance-2"]
- `PackageURL`: 包的下载地址，必须是HTTPS
- `DeployID`: 发布任务唯一标识
- `Timeout`: 单个实例发布超时时间
- `RetryCount`: 失败重试次数

**返回结果**:
```go
type DeployResult struct {
    DeployID       string    `json:"deploy_id"`
    Service        string    `json:"service"`
    Version        string    `json:"version"`
    Status         string    `json:"status"`
    TotalInstances int       `json:"total_instances"`
    StartedAt      time.Time `json:"started_at"`
}
```

### 1.2 查询发布状态

**函数描述**: 查询指定发布任务的执行状态

**函数签名**:
```go
func GetDeploymentStatus(deployID string) (*DeployStatus, error)
```

**输入参数**:
```go
deployID string // 发布任务ID
```

**返回结果**:
```go
type DeployStatus struct {
    DeployID  string `json:"deploy_id"`
    Service   string `json:"service"`
    Version   string `json:"version"`
    Status    string `json:"status"`
    Progress  struct {
        Total     int `json:"total"`
        Completed int `json:"completed"`
        Failed    int `json:"failed"`
        Pending   int `json:"pending"`
    } `json:"progress"`
    Instances []struct {
        InstanceID     string    `json:"instance_id"`
        Status         string    `json:"status"`
        CurrentVersion string    `json:"current_version"`
        TargetVersion  string    `json:"target_version"`
        StartedAt      time.Time `json:"started_at"`
        CompletedAt    *time.Time `json:"completed_at"`
        ErrorMessage   string    `json:"error_message"`
    } `json:"instances"`
    StartedAt time.Time `json:"started_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

**状态说明**:
- `pending`: 等待执行
- `in_progress`: 执行中
- `completed`: 执行完成
- `failed`: 执行失败
- `cancelled`: 已取消

### 1.3 取消发布任务

**函数描述**: 取消正在执行的发布任务

**函数签名**:
```go
func CancelDeployment(deployID string) (*CancelResult, error)
```

**输入参数**:
```go
deployID string // 发布任务ID
```

**返回结果**:
```go
type CancelResult struct {
    DeployID    string    `json:"deploy_id"`
    Status      string    `json:"status"`
    CancelledAt time.Time `json:"cancelled_at"`
}
```

## 2. 版本查询接口

### 2.1 获取服务所有实例的运行版本

**函数描述**: 获取指定服务所有实例的当前运行版本信息

**函数签名**:
```go
func GetServiceInstanceVersions(serviceName string, includeStopped bool) (*ServiceVersions, error)
```

**输入参数**:
```go
serviceName    string // 服务名称
includeStopped bool   // 是否包含停止的实例，默认false
```

**返回结果**:
```go
type ServiceVersions struct {
    Service string `json:"service"`
    Instances []struct {
        InstanceID  string    `json:"instance_id"`
        Version     string    `json:"version"`
        Status      string    `json:"status"`
        LastUpdated time.Time `json:"last_updated"`
    } `json:"instances"`
    VersionSummary  map[string]int `json:"version_summary"`
    TotalInstances  int            `json:"total_instances"`
    UpdatedAt       time.Time      `json:"updated_at"`
}
```

### 2.2 获取服务的实例列表

**函数描述**: 获取指定服务的所有实例详细信息

**函数签名**:
```go
func GetServiceInstances(params *InstanceQueryParams) (*ServiceInstances, error)
```

**输入参数**:
```go
type InstanceQueryParams struct {
    ServiceName string `json:"service_name"` // 服务名称
    Status      string `json:"status"`       // 实例状态过滤，可选值：running, stopped, error
    Version     string `json:"version"`      // 版本过滤
    Limit       int    `json:"limit"`        // 返回数量限制，默认100
    Offset      int    `json:"offset"`       // 偏移量，默认0
}
```

**返回结果**:
```go
type ServiceInstances struct {
    Service string `json:"service"`
    Instances []struct {
        InstanceID    string            `json:"instance_id"`
        Host          string            `json:"host"`
        Port          int               `json:"port"`
        Status        string            `json:"status"`
        Version       string            `json:"version"`
        LastHeartbeat time.Time         `json:"last_heartbeat"`
        Metadata      map[string]string `json:"metadata"`
    } `json:"instances"`
    Total  int `json:"total"`
    Limit  int `json:"limit"`
    Offset int `json:"offset"`
}
```

### 2.3 获取实例版本历史

**接口描述**: 获取指定实例的版本变更历史

**请求信息**:
- **URL**: `GET /v1/instance/{instance_id}/version-history`
- **Method**: `GET`

**路径参数**:
- `instance_id`: 实例ID

**查询参数**:
- `limit`: 返回数量限制，默认50
- `offset`: 偏移量，默认0

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "instance_id": "instance-1",
    "service": "user-service",
    "history": [
      {
        "version": "v1.2.3",
        "action": "deploy",
        "deploy_id": "deploy-12345",
        "timestamp": "2024-01-15T10:31:00Z",
        "status": "success"
      },
      {
        "version": "v1.2.2",
        "action": "deploy",
        "deploy_id": "deploy-12344",
        "timestamp": "2024-01-15T09:15:00Z",
        "status": "success"
      }
    ],
    "total": 2
  }
}
```

## 3. 回滚相关接口

### 3.1 单实例回滚

**函数描述**: 对指定实例执行回滚操作

**函数签名**:
```go
func RollbackInstance(params *InstanceRollbackParams) (*RollbackResult, error)
```

**输入参数**:
```go
type InstanceRollbackParams struct {
    InstanceID    string `json:"instance_id"`    // 必填，实例ID
    TargetVersion string `json:"target_version"` // 必填，目标版本号
    PackageURL    string `json:"package_url"`    // 可选，包下载URL
    Force         bool   `json:"force"`          // 可选，是否强制回滚，默认false
    Timeout       int    `json:"timeout"`        // 可选，超时时间（秒），默认300
}
```

**返回结果**:
```go
type RollbackResult struct {
    RollbackID    string    `json:"rollback_id"`
    InstanceID    string    `json:"instance_id"`
    TargetVersion string    `json:"target_version"`
    Status        string    `json:"status"`
    StartedAt     time.Time `json:"started_at"`
}
```

### 3.2 批量实例回滚

**函数描述**: 对多个实例执行批量回滚操作

**函数签名**:
```go
func RollbackBatch(params *BatchRollbackParams) (*BatchRollbackResult, error)
```

**输入参数**:
```go
type BatchRollbackParams struct {
    Service       string   `json:"service"`        // 必填，服务名称
    TargetVersion string   `json:"target_version"` // 必填，目标版本号
    PackageURL    string   `json:"package_url"`    // 可选，包下载URL
    Instances     []string `json:"instances"`      // 必填，需要回滚的实例ID列表
    Force         bool     `json:"force"`          // 可选，是否强制回滚
    Timeout       int      `json:"timeout"`        // 可选，超时时间
}
```

**返回结果**:
```go
type BatchRollbackResult struct {
    RollbackID     string    `json:"rollback_id"`
    Service        string    `json:"service"`
    TargetVersion  string    `json:"target_version"`
    Status         string    `json:"status"`
    Instances      []string  `json:"instances"`
    TotalInstances int       `json:"total_instances"`
    StartedAt      time.Time `json:"started_at"`
}
```

### 3.3 查询回滚状态

**函数描述**: 查询指定回滚任务的执行状态

**函数签名**:
```go
func GetRollbackStatus(rollbackID string) (*RollbackStatus, error)
```

**输入参数**:
```go
rollbackID string // 回滚任务ID
```

**返回结果**:
```go
type RollbackStatus struct {
    RollbackID   string `json:"rollback_id"`
    Service      string `json:"service"`
    TargetVersion string `json:"target_version"`
    RollbackType string `json:"rollback_type"`
    Status       string `json:"status"`
    Progress     struct {
        Total     int `json:"total"`
        Completed int `json:"completed"`
        Failed    int `json:"failed"`
        Pending   int `json:"pending"`
    } `json:"progress"`
    Instances []struct {
        InstanceID     string     `json:"instance_id"`
        Status         string     `json:"status"`
        CurrentVersion string     `json:"current_version"`
        TargetVersion  string     `json:"target_version"`
        StartedAt      time.Time  `json:"started_at"`
        CompletedAt    *time.Time `json:"completed_at"`
        ErrorMessage   string     `json:"error_message"`
    } `json:"instances"`
    StartedAt   time.Time  `json:"started_at"`
    CompletedAt *time.Time `json:"completed_at"`
}
```

### 3.4 取消回滚任务

**函数描述**: 取消正在执行的回滚任务

**函数签名**:
```go
func CancelRollback(rollbackID string) (*CancelResult, error)
```

**输入参数**:
```go
rollbackID string // 回滚任务ID
```

**返回结果**:
```go
type CancelResult struct {
    RollbackID  string    `json:"rollback_id"`
    Status      string    `json:"status"`
    CancelledAt time.Time `json:"cancelled_at"`
}
```

## 4. 系统管理接口

### 4.1 健康检查

**函数描述**: 检查发布系统健康状态

**函数签名**:
```go
func GetSystemHealth() (*SystemHealth, error)
```

**返回结果**:
```go
type SystemHealth struct {
    Status    string    `json:"status"`
    Version   string    `json:"version"`
    Uptime    string    `json:"uptime"`
    Database  string    `json:"database"`
    Cache     string    `json:"cache"`
    Timestamp time.Time `json:"timestamp"`
}
```

### 4.2 获取系统统计信息

**函数描述**: 获取发布系统的统计信息

**函数签名**:
```go
func GetSystemStats(period string) (*SystemStats, error)
```

**输入参数**:
```go
period string // 统计周期，可选值：1h, 24h, 7d, 30d，默认24h
```

**返回结果**:
```go
type SystemStats struct {
    Period string `json:"period"`
    Deployments struct {
        Total       int     `json:"total"`
        Success     int     `json:"success"`
        Failed      int     `json:"failed"`
        SuccessRate float64 `json:"success_rate"`
    } `json:"deployments"`
    Rollbacks struct {
        Total       int     `json:"total"`
        Success     int     `json:"success"`
        Failed      int     `json:"failed"`
        SuccessRate float64 `json:"success_rate"`
    } `json:"rollbacks"`
    Instances struct {
        Total     int `json:"total"`
        Healthy   int `json:"healthy"`
        Unhealthy int `json:"unhealthy"`
    } `json:"instances"`
    Timestamp time.Time `json:"timestamp"`
}
```

## 5. 使用示例

### 完整发布流程示例

```go
package main

import (
    "fmt"
    "log"
)

func main() {
    // 1. 触发发布
    deployParams := &DeployParams{
        Service:    "user-service",
        Version:    "v1.2.3",
        Instances:  []string{"instance-1", "instance-2"},
        PackageURL: "https://packages.example.com/user-service/v1.2.3.tar.gz",
        DeployID:   "deploy-12345",
        Timeout:    300,
        RetryCount: 3,
    }
    
    result, err := ExecuteDeployment(deployParams)
    if err != nil {
        log.Fatalf("发布失败: %v", err)
    }
    fmt.Printf("发布启动成功: %s\n", result.DeployID)
    
    // 2. 查询发布状态
    status, err := GetDeploymentStatus("deploy-12345")
    if err != nil {
        log.Fatalf("查询状态失败: %v", err)
    }
    fmt.Printf("发布状态: %s, 进度: %d/%d\n", 
        status.Status, status.Progress.Completed, status.Progress.Total)
    
    // 3. 如果发布失败，执行回滚
    if status.Status == "failed" {
        rollbackParams := &BatchRollbackParams{
            Service:       "user-service",
            TargetVersion: "v1.2.2",
            PackageURL:    "https://packages.example.com/user-service/v1.2.2.tar.gz",
            Instances:     []string{"instance-1", "instance-2"},
        }
        
        rollbackResult, err := RollbackBatch(rollbackParams)
        if err != nil {
            log.Fatalf("回滚失败: %v", err)
        }
        fmt.Printf("回滚启动成功: %s\n", rollbackResult.RollbackID)
    }
}
```
