# 发布系统接口参考文档

## 基础信息

本文档描述发布系统的Go接口设计，采用面向接口编程的方式，将功能按职责划分为多个接口。

## 1. 接口概览

发布系统提供以下外部接口：

- **DeployService**: 发布服务接口，负责发布和回滚操作的执行
- **InstanceManager**: 实例管理接口，负责实例信息查询和状态管理

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
    DeployID       string    `json:"deploy_id"`       // 发布任务ID
    Service        string    `json:"service"`         // 服务名称
    Version        string    `json:"version"`         // 发布的目标版本
    Message        string    `json:"message"`         // 发布完成状态描述
    Instances      []string  `json:"instances"`       // 实际发布的实例ID列表
    TotalInstances int       `json:"total_instances"` // 发布的实例总数
    CompletedAt    time.Time `json:"completed_at"`    // 发布完成时间
}
```

**字段说明**:
- `DeployID`: 发布任务的唯一标识符，与请求参数中的DeployID一致
- `Service`: 发布的服务名称
- `Version`: 成功发布的版本号
- `Message`: 发布操作的完成状态描述，如"发布成功"、"部分实例发布失败"等
- `Instances`: 实际参与发布的实例ID列表，可能与请求中的实例列表有差异
- `TotalInstances`: 实际发布的实例数量
- `CompletedAt`: 发布操作完成的时间戳

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
    DeployID    string    `json:"deploy_id"`    // 被取消的发布任务ID
    Message     string    `json:"message"`      // 取消操作状态描述
    CancelledAt time.Time `json:"cancelled_at"` // 取消操作完成时间
}
```

**字段说明**:
- `DeployID`: 被取消的发布任务唯一标识符
- `Message`: 取消操作的状态描述，如"发布已成功取消"、"发布已完成无法取消"等
- `CancelledAt`: 取消操作完成的时间戳

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
- `PackageURL`: 包的下载地址，HTTP或者本地路径

**返回结果**:
```go
type RollbackResult struct {
    RollbackID     string    `json:"rollback_id"`     // 回滚任务ID
    Service        string    `json:"service"`         // 服务名称
    TargetVersion  string    `json:"target_version"`  // 回滚的目标版本
    Message        string    `json:"message"`         // 回滚完成状态描述
    Instances      []string  `json:"instances"`       // 实际回滚的实例ID列表
    TotalInstances int       `json:"total_instances"` // 回滚的实例总数
    CompletedAt    time.Time `json:"completed_at"`    // 回滚完成时间
}
```

**字段说明**:
- `RollbackID`: 回滚任务的唯一标识符，与请求参数中的RollbackID一致
- `Service`: 回滚的服务名称
- `TargetVersion`: 成功回滚到的目标版本号
- `Message`: 回滚操作的完成状态描述，如"回滚成功"、"部分实例回滚失败"等
- `Instances`: 实际参与回滚的实例ID列表，可能与请求中的实例列表有差异
- `TotalInstances`: 实际回滚的实例数量
- `CompletedAt`: 回滚操作完成的时间戳

## 3. InstanceManager接口

### 3.1 接口定义

实例管理接口，负责实例信息查询和状态管理，发布模块和服务管理模块都需要使用。

```go
type InstanceManager interface {
    GetServiceInstances(serviceName string) ([]string, error)
    GetInstancesInfo(instanceIDs []string) (map[string]*InstanceInfo, error)
    GetInstancesVersion(instanceIDs []string) (map[string]string, error)
    GetInstanceVersionHistory(instanceID string) ([]*VersionInfo, error)
    CheckInstanceHealth(instanceIDs []string) (map[string]*HealthStatus, error)
}
```

### 3.2 数据结构定义

**InstanceInfo结构体**:
```go
type InstanceInfo struct {
    InstanceID  string `json:"instance_id"`  // 实例唯一标识符
    ServiceName string `json:"service_name"` // 所属服务名称
    Version     string `json:"version"`      // 当前运行的版本号
    Status      string `json:"status"`       // 实例运行状态
}
```

**字段说明**:
- `InstanceID`: 实例的全局唯一标识符，如"user-service-001"
- `ServiceName`: 实例所属的服务名称，如"user-service"
- `Version`: 实例当前运行的软件版本号，如"v1.2.3"
- `Status`: 实例的运行状态，如"running"、"stopped"、"starting"等

**HealthStatus结构体**:
```go
type HealthStatus struct {
    InstanceID string    `json:"instance_id"`           // 实例唯一标识符
    IsHealthy  bool      `json:"is_healthy"`            // 健康检查结果
    CheckedAt  time.Time `json:"checked_at"`            // 健康检查时间
    Message    string    `json:"message,omitempty"`     // 健康状态描述信息
}
```

**字段说明**:
- `InstanceID`: 被检查实例的唯一标识符
- `IsHealthy`: 健康检查结果，true表示健康，false表示不健康
- `CheckedAt`: 执行健康检查的时间戳
- `Message`: 健康状态的详细描述信息，如"HTTP 200 响应正常"、"连接超时"等

**VersionInfo结构体**:
```go
type VersionInfo struct {
    Version    string    `json:"version"`     // 版本号
    DeployedAt time.Time `json:"deployed_at"` // 发布时间
    DeployID   string    `json:"deploy_id"`   // 发布任务ID
    Status     string    `json:"status"`      // 发布方式状态
}
```

**字段说明**:
- `Version`: 版本号，如"v1.2.3"
- `DeployedAt`: 该版本发布到实例的时间戳
- `DeployID`: 执行发布的任务ID，用于追溯发布来源
- `Status`: 发布方式状态，"deploy"表示通过正常发布，"rollback"表示通过回滚发布

### 3.3 GetServiceInstances方法

**方法描述**: 获取指定服务的所有实例列表

**方法签名**:
```go
GetServiceInstances(serviceName string) ([]string, error)
```

**输入参数**:
```go
serviceName string // 服务名称
```

**返回结果**: `[]string` - 实例ID数组

### 3.4 GetInstancesInfo方法

**方法描述**: 批量获取多个实例的详细信息

**方法签名**:
```go
GetInstancesInfo(instanceIDs []string) (map[string]*InstanceInfo, error)
```

**输入参数**:
```go
instanceIDs []string // 实例ID数组
```

**返回结果**: `map[string]*InstanceInfo` - 实例ID到实例信息的映射

### 3.7 CheckInstanceHealth方法

**方法描述**: 检查实例的健康状态，支持单个或多个实例

**方法签名**:
```go
CheckInstanceHealth(instanceIDs []string) (map[string]*HealthStatus, error)
```

**输入参数**:
```go
instanceIDs []string // 实例ID数组
```

**返回结果**: `map[string]*HealthStatus` - 实例ID到健康状态的映射

### 3.5 GetInstancesVersion方法

**方法描述**: 批量获取多个实例的当前版本

**方法签名**:
```go
GetInstancesVersion(instanceIDs []string) (map[string]string, error)
```

**输入参数**:
```go
instanceIDs []string // 实例ID数组
```

**返回结果**: `map[string]string` - 实例ID到版本号的映射

### 3.6 GetInstanceVersionHistory方法

**方法描述**: 获取指定实例的版本历史记录

**方法签名**:
```go
GetInstanceVersionHistory(instanceID string) ([]*VersionInfo, error)
```

**输入参数**:
```go
instanceID string // 实例ID
```

**返回结果**: `[]*VersionInfo` - 版本历史数组

## 4. 使用示例

### 4.1 接口实现示例

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

### 4.2 完整发布流程示例

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

## 5. 内部工具函数

### 5.1 ValidatePackageURL函数

**函数描述**: 验证包URL的有效性和安全性

**函数签名**:
```go
func ValidatePackageURL(packageURL string) error
```

**输入参数**:
```go
packageURL string // 包下载URL
```

**返回结果**: `error` - 验证失败时返回错误信息

**验证规则**:
- URL必须使用HTTPS协议
- URL格式必须正确
- 域名必须在白名单中（可选）
- 文件扩展名必须符合要求（如.tar.gz, .zip等）

**使用示例**:
```go
func (fd *floyDeployService) ExecuteDeployment(params *DeployParams) (*DeployResult, error) {
    // 验证包URL
    if err := ValidatePackageURL(params.PackageURL); err != nil {
        return nil, fmt.Errorf("无效的包URL: %v", err)
    }
    
    // 继续执行发布逻辑...
}
```
