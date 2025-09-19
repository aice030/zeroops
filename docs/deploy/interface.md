# 发布系统接口参考文档

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

### 2.3 ExecuteRollback方法

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

## 3. InstanceManager接口

### 3.1 接口定义

实例管理接口，负责实例信息查询和状态管理，发布模块和服务管理模块都需要使用。

```go
type InstanceManager interface {
    GetServiceInstances(serviceName string, version ...string) ([]*InstanceInfo, error)
    GetInstanceVersionHistory(instanceID string) ([]*VersionInfo, error)
}
```

### 3.2 数据结构定义

**InstanceInfo结构体**:
```go
type InstanceInfo struct {
    InstanceID  string `json:"instance_id"`  // 实例唯一标识符
    ServiceName string `json:"service_name"` // 所属服务名称
    Version     string `json:"version"`      // 当前运行的版本号
    Status      string `json:"status"`       // 实例运行状态 - 'active'运行中；'pending'发布中；'error'出现故障
}
```

**VersionInfo结构体**:
```go
type VersionInfo struct {
    Version string `json:"version"` // 版本号
    Status  string `json:"status"`  // 版本状态 - 'acitve'当前运行版本；'stable'稳定版本；'deprecated'已废弃版本
}
```

### 3.3 GetServiceInstances方法

**方法描述**: 获取指定服务的实例详细信息，可选择按版本过滤

**方法签名**:
```go
GetServiceInstances(serviceName string, version ...string) ([]*InstanceInfo, error)
```

**输入参数**:
```go
serviceName string   // 必填，服务名称
version     ...string // 选填，指定版本号进行过滤，未输入则默认获取全部版本的运行实例
```

**返回结果**: `[]*InstanceInfo` - 实例信息数组

### 3.4 GetInstanceVersionHistory方法

**方法描述**: 获取指定实例的版本历史记录

**方法签名**:
```go
GetInstanceVersionHistory(instanceID string) ([]*VersionInfo, error)
```

**输入参数**:
```go
instanceID string // 必填，实例ID
```

**返回结果**: `[]*VersionInfo` - 版本历史数组

## 4. 内部工具函数

### 4.1 ValidatePackageURL函数

**函数描述**: 验证是否能通过URL找到包

**函数签名**:
```go
func ValidatePackageURL(packageURL string) error
```

**输入参数**:
```go
packageURL string // 必填，包下载URL
```

**返回结果**: `error` - 验证失败时返回错误信息

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

### 4.2 GetInstanceHost函数

**函数描述**: 根据实例ID获取实例的IP地址

**函数签名**:
```go
func GetInstanceHost(instanceID string) (string, error)
```

**输入参数**:
```go
instanceID string // 必填，实例ID
```

**返回结果**: `string` - 实例的IP地址，获取失败时返回错误信息

### 4.3 GetInstancePort函数

**函数描述**: 根据服务名和实例IP获取实例的端口号

**函数签名**:
```go
func GetInstancePort(serviceName, instanceHost string) (int, error)
```

**输入参数**:
```go
serviceName  string // 必填，服务名称
instanceHost string // 必填，实例IP地址
```

**返回结果**: `int` - 实例的端口号，获取失败时返回错误信息

### 4.4 CheckInstanceHealth函数

**函数描述**: 检查单个实例是否有响应，用于发布前验证目标实例的可用性

**函数签名**:
```go
func CheckInstanceHealth(instanceID string) (bool, error)
```

**输入参数**:
```go
instanceID string // 必填，实例ID
```

**返回结果**: `bool` - 健康检查结果，true表示实例有响应，false表示无响应
