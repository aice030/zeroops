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

### 3.1 发布相关接口

#### 3.1.1 触发发布

**函数签名：**
```go
func ExecuteDeployment(params *DeployParams) (*DeployResult, error)
```

**输入参数：**
```go
type DeployParams struct {
    Service    string   `json:"service"`     // 服务名称
    Version    string   `json:"version"`     // 目标版本号
    Instances  []string `json:"instances"`   // 实例ID列表
    PackageURL string   `json:"package_url"` // 包下载URL
    DeployID   string   `json:"deploy_id"`   // 发布任务ID
    Timeout    int      `json:"timeout"`     // 超时时间（秒）
    RetryCount int      `json:"retry_count"` // 重试次数
}
```

**返回结果：**
```go
type DeployResult struct {
    DeployID  string    `json:"deploy_id"`
    Status    string    `json:"status"`
    StartedAt time.Time `json:"started_at"`
}
```

#### 3.1.2 查询发布状态

**函数签名：**
```go
func GetDeploymentStatus(deployID string) (*DeployStatus, error)
```

**输入参数：**
```go
deployID string // 发布任务ID
```

**返回结果：**
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
        InstanceID string    `json:"instance_id"`
        Status     string    `json:"status"`
        Version    string    `json:"version"`
        UpdatedAt  time.Time `json:"updated_at"`
    } `json:"instances"`
    StartedAt time.Time `json:"started_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### 3.2 版本查询接口

#### 3.2.1 获取服务所有实例的运行版本

**函数签名：**
```go
func GetServiceInstanceVersions(serviceName string) (*ServiceVersions, error)
```

**输入参数：**
```go
serviceName string // 服务名称
```

**返回结果：**
```go
type ServiceVersions struct {
    Service string `json:"service"`
    Instances []struct {
        InstanceID  string    `json:"instance_id"`
        Version     string    `json:"version"`
        Status      string    `json:"status"`
        LastUpdated time.Time `json:"last_updated"`
    } `json:"instances"`
    VersionSummary map[string]int `json:"version_summary"`
}
```

#### 3.2.2 获取服务的实例列表

**函数签名：**
```go
func GetServiceInstances(serviceName string) (*ServiceInstances, error)
```

**输入参数：**
```go
serviceName string // 服务名称
```

**返回结果：**
```go
type ServiceInstances struct {
    Service string `json:"service"`
    Instances []struct {
        InstanceID string `json:"instance_id"`
        Host       string `json:"host"`
        Port       int    `json:"port"`
        Status     string `json:"status"`
        Version    string `json:"version"`
    } `json:"instances"`
    Total int `json:"total"`
}
```

### 3.3 回滚相关接口

#### 3.3.1 单实例回滚

**函数签名：**
```go
func RollbackInstance(params *InstanceRollbackParams) (*RollbackResult, error)
```

**输入参数：**
```go
type InstanceRollbackParams struct {
    InstanceID    string `json:"instance_id"`    // 实例ID
    TargetVersion string `json:"target_version"` // 目标版本
    PackageURL    string `json:"package_url"`    // 包下载URL
}
```

**返回结果：**
```go
type RollbackResult struct {
    RollbackID    string `json:"rollback_id"`
    InstanceID    string `json:"instance_id"`
    TargetVersion string `json:"target_version"`
    Status        string `json:"status"`
}
```

#### 3.3.2 批量实例回滚

**函数签名：**
```go
func RollbackBatch(params *BatchRollbackParams) (*BatchRollbackResult, error)
```

**输入参数：**
```go
type BatchRollbackParams struct {
    Service       string   `json:"service"`        // 服务名称
    TargetVersion string   `json:"target_version"` // 目标版本
    PackageURL    string   `json:"package_url"`    // 包下载URL
    Instances     []string `json:"instances"`      // 实例ID列表
}
```

**返回结果：**
```go
type BatchRollbackResult struct {
    RollbackID    string   `json:"rollback_id"`
    Service       string   `json:"service"`
    TargetVersion string   `json:"target_version"`
    Status        string   `json:"status"`
    Instances     []string `json:"instances"`
}
```

#### 3.3.3 查询回滚状态

**函数签名：**
```go
func GetRollbackStatus(rollbackID string) (*RollbackStatus, error)
```

**输入参数：**
```go
rollbackID string // 回滚任务ID
```

**返回结果：**
```go
type RollbackStatus struct {
    RollbackID    string `json:"rollback_id"`
    Service       string `json:"service"`
    TargetVersion string `json:"target_version"`
    Status        string `json:"status"`
    Instances     []struct {
        InstanceID string    `json:"instance_id"`
        Status     string    `json:"status"`
        Version    string    `json:"version"`
        UpdatedAt  time.Time `json:"updated_at"`
    } `json:"instances"`
    StartedAt   time.Time `json:"started_at"`
    CompletedAt time.Time `json:"completed_at"`
}
```

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
