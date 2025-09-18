# 发布系统接口参考文档

## 基础信息

本文档描述发布系统的Go接口设计，采用面向接口编程的方式，将功能按职责划分为多个接口。

## 1. 接口概览

发布系统按职责划分为以下核心接口：

- **DeployExecutor**: 发布执行接口，负责发布任务的执行
- **VersionManager**: 版本管理接口，负责服务实例版本信息的查询
- **RollbackManager**: 回滚管理接口，负责回滚操作的执行

## 2. DeployExecutor接口

### 2.1 接口定义

发布执行接口，负责发布任务的执行。

```go
type DeployExecutor interface {
    ExecuteDeployment(params *DeployParams) (*DeployResult, error)
    GetDeploymentStatus(deployID string) (*DeployStatus, error)
    CancelDeployment(deployID string) (*CancelResult, error)
}
```

### 2.2 ExecuteDeployment方法

**方法描述**: 触发指定服务版本的发布操作

**方法签名**:
```go
ExecuteDeployment(params *DeployParams) (*DeployResult, error)
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

### 2.3 GetDeploymentStatus方法

**方法描述**: 查询指定发布任务的执行状态

**方法签名**:
```go
GetDeploymentStatus(deployID string) (*DeployStatus, error)
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
        InstanceID     string     `json:"instance_id"`
        CurrentVersion string     `json:"current_version"`
        TargetVersion  string     `json:"target_version"`
        StartedAt      time.Time  `json:"started_at"`
        CompletedAt    *time.Time `json:"completed_at"`
        ErrorMessage   string     `json:"error_message"`
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

### 2.4 CancelDeployment方法

**方法描述**: 取消正在执行的发布任务

**方法签名**:
```go
CancelDeployment(deployID string) (*CancelResult, error)
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

## 3. VersionManager接口

### 3.1 接口定义

版本管理接口，负责服务实例版本信息的查询。

```go
type VersionManager interface {
    GetServiceInstanceVersions(serviceName string) (*ServiceVersions, error)
    GetServiceInstances(params *InstanceQueryParams) (*ServiceInstances, error)
}
```

### 3.2 GetServiceInstanceVersions方法

**方法描述**: 获取指定服务所有实例的当前运行版本信息

**方法签名**:
```go
GetServiceInstanceVersions(serviceName string) (*ServiceVersions, error)
```

**输入参数**:
```go
serviceName string // 服务名称
```

**返回结果**:
```go
type ServiceVersions struct {
    Service string `json:"service"`
    Instances []struct {
        InstanceID  string    `json:"instance_id"`
        Version     string    `json:"version"`
        LastUpdated time.Time `json:"last_updated"`
    } `json:"instances"`
    VersionSummary  map[string]int `json:"version_summary"`
    TotalInstances  int            `json:"total_instances"`
    UpdatedAt       time.Time      `json:"updated_at"`
}
```

### 3.3 GetServiceInstances方法

**方法描述**: 获取指定服务的所有实例详细信息

**方法签名**:
```go
GetServiceInstances(params *InstanceQueryParams) (*ServiceInstances, error)
```

**输入参数**:
```go
type InstanceQueryParams struct {
    ServiceName string `json:"service_name"` // 服务名称
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
        Version       string            `json:"version"`
        LastHeartbeat time.Time         `json:"last_heartbeat"`
        Metadata      map[string]string `json:"metadata"`
    } `json:"instances"`
    Total  int `json:"total"`
    Limit  int `json:"limit"`
    Offset int `json:"offset"`
}
```


## 4. RollbackManager接口

### 4.1 接口定义

回滚管理接口，负责回滚操作的执行。

```go
type RollbackManager interface {
    RollbackInstance(params *InstanceRollbackParams) (*RollbackResult, error)
    RollbackBatch(params *BatchRollbackParams) (*BatchRollbackResult, error)
}
```

### 4.2 RollbackInstance方法

**方法描述**: 对指定实例执行回滚操作

**方法签名**:
```go
RollbackInstance(params *InstanceRollbackParams) (*RollbackResult, error)
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

### 4.3 RollbackBatch方法

**方法描述**: 对多个实例执行批量回滚操作

**方法签名**:
```go
RollbackBatch(params *BatchRollbackParams) (*BatchRollbackResult, error)
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



## 5. 统一服务接口

如果需要一个统一的服务接口，可以组合所有功能接口：

```go
type DeployService interface {
    DeployExecutor
    VersionManager
    RollbackManager
}
```

## 6. 使用示例

### 6.1 接口实现示例

```go
// 实现结构体
type floyDeployService struct {
    logger   Logger
    executor Executor
    database Database
}

// 实现DeployExecutor接口
func (fd *floyDeployService) ExecuteDeployment(params *DeployParams) (*DeployResult, error) {
    // 实现逻辑
}

func (fd *floyDeployService) GetDeploymentStatus(deployID string) (*DeployStatus, error) {
    // 实现逻辑
}

func (fd *floyDeployService) CancelDeployment(deployID string) (*CancelResult, error) {
    // 实现逻辑
}

// 实现VersionManager接口
func (fd *floyDeployService) GetServiceInstanceVersions(serviceName string) (*ServiceVersions, error) {
    // 实现逻辑
}

func (fd *floyDeployService) GetServiceInstances(params *InstanceQueryParams) (*ServiceInstances, error) {
    // 实现逻辑
}

// 实现RollbackManager接口
func (fd *floyDeployService) RollbackInstance(params *InstanceRollbackParams) (*RollbackResult, error) {
    // 实现逻辑
}

func (fd *floyDeployService) RollbackBatch(params *BatchRollbackParams) (*BatchRollbackResult, error) {
    // 实现逻辑
}

// 构造函数
func NewDeployService(logger Logger, executor Executor, database Database) DeployService {
    return &floyDeployService{
        logger:   logger,
        executor: executor,
        database: database,
    }
}
```

### 6.2 完整发布流程示例

```go
package main

import (
    "fmt"
    "log"
)

func main() {
    // 初始化服务
    floyDeployService := NewDeployService(logger, executor, database)
    
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
    
    result, err := floyDeployService.ExecuteDeployment(deployParams)
    if err != nil {
        log.Fatalf("发布失败: %v", err)
    }
    fmt.Printf("发布启动成功: %s\n", result.DeployID)
    
    // 2. 查询服务实例版本
    versions, err := floyDeployService.GetServiceInstanceVersions("user-service")
    if err != nil {
        log.Fatalf("查询版本失败: %v", err)
    }
    fmt.Printf("服务实例版本: %+v\n", versions.VersionSummary)
    
    // 3. 如果需要，执行回滚操作
    rollbackParams := &BatchRollbackParams{
        Service:       "user-service",
        TargetVersion: "v1.2.2",
        PackageURL:    "https://packages.example.com/user-service/v1.2.2.tar.gz",
        Instances:     []string{"instance-1", "instance-2"},
    }
    
    rollbackResult, err := floyDeployService.RollbackBatch(rollbackParams)
    if err != nil {
        log.Fatalf("回滚失败: %v", err)
    }
    fmt.Printf("回滚启动成功: %s\n", rollbackResult.RollbackID)
}
```
