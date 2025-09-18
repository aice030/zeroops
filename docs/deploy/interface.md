# 发布系统接口参考文档

## 基础信息

本文档描述发布系统的Go接口设计，采用面向接口编程的方式，将功能按职责划分为多个接口。

## 1. 接口概览

发布系统提供统一的外部接口：

- **DeployService**: 发布服务接口，负责发布和回滚操作的执行

## 2. DeployService接口

### 2.1 接口定义

发布服务接口，负责发布和回滚操作的执行。

```go
type DeployService interface {
    ExecuteDeployment(params *DeployParams) (*DeployResult, error)
    CancelDeployment(deployID string) (*CancelResult, error)
    ExecuteRollback(params *RollbackParams) (*RollbackResult, error)
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
    DeployID   string   `json:"deploy_id"`   // 必填，发布任务ID
    Service    string   `json:"service"`     // 必填，服务名称
    Version    string   `json:"version"`     // 必填，目标版本号
    Instances  []string `json:"instances"`   // 必填，实例ID列表
    PackageURL string   `json:"package_url"` // 必填，包下载URL
}
```

**参数说明**:
- `DeployID`: 发布任务唯一标识
- `Service`: 服务名称，如 "user-service"
- `Version`: 版本号，如 "v1.2.3"
- `Instances`: 实例ID数组，如 ["instance-1", "instance-2"]
- `PackageURL`: 包的下载地址，必须是HTTPS

**返回结果**:
```go
type DeployResult struct {
    DeployID       string    `json:"deploy_id"`
    Service        string    `json:"service"`
    Version        string    `json:"version"`
    Message        string    `json:"message"`
    Instances      []string  `json:"instances"`
    TotalInstances int       `json:"total_instances"`
    CompletedAt    time.Time `json:"completed_at"`
}
```

### 2.3 CancelDeployment方法

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
    Message     string    `json:"message"`
    CancelledAt time.Time `json:"cancelled_at"`
}
```

### 2.4 ExecuteRollback方法

**方法描述**: 对指定实例执行回滚操作，支持单实例或批量实例回滚

**方法签名**:
```go
ExecuteRollback(params *RollbackParams) (*RollbackResult, error)
```

**输入参数**:
```go
type RollbackParams struct {
    RollbackID    string   `json:"rollback_id"`    // 必填，回滚任务ID
    Service       string   `json:"service"`        // 必填，服务名称
    TargetVersion string   `json:"target_version"` // 必填，目标版本号
    Instances     []string `json:"instances"`      // 必填，实例ID列表
    PackageURL    string   `json:"package_url"`    // 必填，包下载URL
}
```

**参数说明**:
- `RollbackID`: 回滚任务唯一标识
- `Service`: 服务名称，如 "user-service"
- `TargetVersion`: 目标版本号，如 "v1.2.2"
- `Instances`: 实例ID数组，单实例回滚传入一个元素，批量回滚传入多个元素
- `PackageURL`: 包的下载地址，必须是HTTPS

**返回结果**:
```go
type RollbackResult struct {
    RollbackID     string    `json:"rollback_id"`
    Service        string    `json:"service"`
    TargetVersion  string    `json:"target_version"`
    Message        string    `json:"message"`
    Instances      []string  `json:"instances"`
    TotalInstances int       `json:"total_instances"`
    CompletedAt    time.Time `json:"completed_at"`
}
```


## 3. 使用示例

### 3.1 接口实现示例

```go
// 实现结构体
type floyDeployService struct {
    logger   Logger
    executor Executor
    database Database
}

// 实现DeployService接口
func (fd *floyDeployService) ExecuteDeployment(params *DeployParams) (*DeployResult, error) {
    // 实现逻辑
}

func (fd *floyDeployService) CancelDeployment(deployID string) (*CancelResult, error) {
    // 实现逻辑
}

func (fd *floyDeployService) ExecuteRollback(params *RollbackParams) (*RollbackResult, error) {
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

### 3.2 完整发布流程示例

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
        DeployID:   "deploy-12345",
        Service:    "user-service",
        Version:    "v1.2.3",
        Instances:  []string{"instance-1", "instance-2"},
        PackageURL: "https://packages.example.com/user-service/v1.2.3.tar.gz",
    }
    
    result, err := floyDeployService.ExecuteDeployment(deployParams)
    if err != nil {
        log.Fatalf("发布失败: %v", err)
    }
    fmt.Printf("发布启动成功: %s\n", result.DeployID)
    
    // 2. 如果需要，执行回滚操作
    rollbackParams := &RollbackParams{
        RollbackID:    "rollback-67890",
        Service:       "user-service",
        TargetVersion: "v1.2.2",
        Instances:     []string{"instance-1", "instance-2"},
        PackageURL:    "https://packages.example.com/user-service/v1.2.2.tar.gz",
    }
    
    rollbackResult, err := floyDeployService.ExecuteRollback(rollbackParams)
    if err != nil {
        log.Fatalf("回滚失败: %v", err)
    }
    fmt.Printf("回滚启动成功: %s\n", rollbackResult.RollbackID)
}
```
