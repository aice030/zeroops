# 发布系统设计文档

## 1. 系统概述

### 1.1 设计目标
发布系统是一个专注于执行发布动作的轻量级系统，负责：
- 接收调度系统的发布指令
- 执行具体的发布操作
- 管理服务实例的版本状态
- 提供回滚功能

### 1.2 设计原则
- **单一职责**：只负责发布动作的执行，不涉及调度逻辑
- **简单可靠**：提供稳定、可预测的发布操作
- **状态透明**：实时反馈发布状态和实例版本信息
- **快速回滚**：支持单实例和批量回滚操作

### 1.3 系统边界
- **负责**：发布动作执行、版本管理、回滚操作
- **不负责**：发布调度、批次规划、实例选择、健康检查

## 2. 核心功能模块

### 2.1 发布执行模块
负责执行具体的发布操作，包括：
- 接收发布请求（服务名、版本、实例列表）
- 执行发布动作
- 更新实例版本状态
- 记录发布日志

### 2.2 版本管理模块
管理服务版本和实例版本信息：
- 获取服务所有实例的运行版本
- 获取服务的实例列表
- 版本状态查询和更新

### 2.3 回滚模块
提供回滚功能：
- 单实例回滚（支持远程包下载）
- 批量实例回滚（按服务名+版本）
- 回滚状态跟踪

### 2.4 状态管理模块
维护发布和实例状态：
- 发布任务状态管理
- 实例版本状态同步
- 状态变更日志记录

## 3. 接口设计

发布系统提供以下外部接口：

- **DeployService**: 发布服务接口，负责发布和回滚操作的执行
- **InstanceManager**: 实例管理接口，负责实例信息查询和状态管理

### 3.1 DeployService接口

发布服务接口，负责发布和回滚操作的执行。

```go
type DeployService interface {
    ExecuteDeployment(params *DeployParams) (*DeployResult, error)
    CancelDeployment(deployID string) (*CancelResult, error)
    ExecuteRollback(params *RollbackParams) (*RollbackResult, error)
}
```

#### 3.1.1 ExecuteDeployment方法

**方法描述**: 触发指定服务版本的发布操作

**方法签名：**
```go
ExecuteDeployment(params *DeployParams) (*DeployResult, error)
```

**输入参数：**
```go
type DeployParams struct {
    DeployID   string   `json:"deploy_id"`   // 发布任务ID
    Service    string   `json:"service"`     // 服务名称
    Version    string   `json:"version"`     // 目标版本号
    Instances  []string `json:"instances"`   // 实例ID列表
    PackageURL string   `json:"package_url"` // 包下载URL
}
```

**返回结果：**
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

#### 3.1.2 CancelDeployment方法

**方法描述**: 取消正在执行的发布任务

**方法签名：**
```go
CancelDeployment(deployID string) (*CancelResult, error)
```

**输入参数：**
```go
deployID string // 发布任务ID
```

**返回结果：**
```go
type CancelResult struct {
    DeployID    string    `json:"deploy_id"`
    Message     string    `json:"message"`
    CancelledAt time.Time `json:"cancelled_at"`
}
```


#### 3.1.3 ExecuteRollback方法

**方法描述**: 对指定实例执行回滚操作，支持单实例或批量实例回滚

**方法签名：**
```go
ExecuteRollback(params *RollbackParams) (*RollbackResult, error)
```

**输入参数：**
```go
type RollbackParams struct {
    RollbackID    string   `json:"rollback_id"`    // 必填，回滚任务ID
    Service       string   `json:"service"`        // 必填，服务名称
    TargetVersion string   `json:"target_version"` // 必填，目标版本号
    Instances     []string `json:"instances"`      // 必填，实例ID列表
    PackageURL    string   `json:"package_url"`    // 必填，包下载URL
}
```

**参数说明：**
- `RollbackID`: 回滚任务唯一标识
- `Service`: 服务名称，如 "user-service"
- `TargetVersion`: 目标版本号，如 "v1.2.2"
- `Instances`: 实例ID数组，单实例回滚传入一个元素，批量回滚传入多个元素
- `PackageURL`: 包的下载地址，HTTP或者本地路径

**返回结果：**
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

### 3.2 InstanceManager接口

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

#### 3.2.1 数据结构定义

**InstanceInfo结构体**:
```go
type InstanceInfo struct {
    InstanceID    string            `json:"instance_id"`
    ServiceName   string            `json:"service_name"`
    Host          string            `json:"host"`
    Port          int               `json:"port"`
    Version       string            `json:"version"`
    Status        string            `json:"status"`
    LastHeartbeat time.Time         `json:"last_heartbeat"`
    Metadata      map[string]string `json:"metadata"`
}
```

**HealthStatus结构体**:
```go
type HealthStatus struct {
    InstanceID string    `json:"instance_id"`
    IsHealthy  bool      `json:"is_healthy"`
    CheckedAt  time.Time `json:"checked_at"`
    Message    string    `json:"message,omitempty"`
}
```

**VersionInfo结构体**:
```go
type VersionInfo struct {
    Version    string    `json:"version"`
    DeployedAt time.Time `json:"deployed_at"`
    DeployID   string    `json:"deploy_id"`
    Status     string    `json:"status"` // deploy, rollback
}
```

#### 3.2.2 方法说明

**GetServiceInstances方法**: 获取指定服务的所有实例ID列表
- 输入: 服务名称
- 返回: 实例ID数组

**GetInstancesInfo方法**: 批量获取多个实例的详细信息
- 输入: 实例ID数组
- 返回: 实例ID到实例信息的映射

**GetInstancesVersion方法**: 批量获取多个实例的当前版本
- 输入: 实例ID数组
- 返回: 实例ID到版本号的映射

**GetInstanceVersionHistory方法**: 获取指定实例的版本历史记录
- 输入: 实例ID
- 返回: 版本历史数组

**CheckInstanceHealth方法**: 检查实例的健康状态，支持单个或多个实例
- 输入: 实例ID数组
- 返回: 实例ID到健康状态的映射

## 4. 系统架构

### 4.1 整体架构
```
┌─────────────────┐    HTTP API    ┌─────────────────┐    网络通信    ┌─────────────────┐
│   调度系统       │ ──────────────▶ │   发布系统       │ ──────────────▶ │   实例节点       │
│                 │                │                 │                │                 │
│  • 发布指令     │                │  • 执行发布操作  │                │  • 服务实例     │
│  • 批次规划     │                │  • 状态跟踪      │                │  • 版本更新     │
│  • 健康检查     │                │  • 回滚操作      │                │  • 状态上报     │
│                 │                │                 │                │  • 健康检查     │
└─────────────────┘                └─────────────────┘                └─────────────────┘
                                             │
                                             │ 数据存储
                                             ▼
                                    ┌─────────────────┐
                                    │   数据库         │
                                    │                 │
                                    │  • 发布任务     │
                                    │  • 实例版本     │
                                    │  • 回滚记录     │
                                    │  • 状态信息     │
                                    └─────────────────┘
```

### 4.2 核心组件

- **发布执行器**: 执行发布操作，更新实例版本
- **版本管理器**: 管理服务版本信息，查询实例版本状态
- **回滚执行器**: 执行回滚操作，处理包下载
- **状态管理器**: 维护任务状态，同步实例状态
