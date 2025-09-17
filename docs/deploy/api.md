# 发布系统 API 接口参考文档

## 基础信息

- **基础URL**: `http://localhost:8080`
- **API版本**: `v1`
- **内容类型**: `application/json`

## 1. 发布相关接口

### 1.1 触发发布

**接口描述**: 触发指定服务版本的发布操作

**请求信息**:
- **URL**: `POST /v1/deploy/execute`
- **Content-Type**: `application/json`

**请求参数**:
```json
{
  "service": "string",           // 必填，服务名称
  "version": "string",           // 必填，目标版本号
  "instances": ["string"],       // 必填，实例ID列表
  "package_url": "string",       // 必填，包下载URL
  "deploy_id": "string",         // 必填，发布任务ID（由调度系统生成）
  "timeout": 300,                // 可选，超时时间（秒），默认300
  "retry_count": 3               // 可选，重试次数，默认3
}
```

**参数说明**:
- `service`: 服务名称，如 "user-service"
- `version`: 版本号，如 "v1.2.3"
- `instances`: 实例ID数组，如 ["instance-1", "instance-2"]
- `package_url`: 包的下载地址，必须是HTTPS
- `deploy_id`: 发布任务唯一标识
- `timeout`: 单个实例发布超时时间
- `retry_count`: 失败重试次数

**响应示例**:
```json
{
  "code": 200,
  "message": "deployment started successfully",
  "data": {
    "deploy_id": "deploy-12345",
    "service": "user-service",
    "version": "v1.2.3",
    "status": "started",
    "total_instances": 2,
    "started_at": "2024-01-15T10:30:00Z"
  }
}
```

**错误码**:
- `40001`: 服务不存在
- `40002`: 版本不存在
- `40003`: 实例不存在
- `40004`: 发布任务ID冲突
- `50001`: 发布启动失败

### 1.2 查询发布状态

**接口描述**: 查询指定发布任务的执行状态

**请求信息**:
- **URL**: `GET /v1/deploy/status/{deploy_id}`
- **Method**: `GET`

**路径参数**:
- `deploy_id`: 发布任务ID

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "deploy_id": "deploy-12345",
    "service": "user-service",
    "version": "v1.2.3",
    "status": "in_progress",
    "progress": {
      "total": 3,
      "completed": 1,
      "failed": 0,
      "pending": 2
    },
    "instances": [
      {
        "instance_id": "instance-1",
        "status": "completed",
        "current_version": "v1.2.3",
        "target_version": "v1.2.3",
        "started_at": "2024-01-15T10:30:00Z",
        "completed_at": "2024-01-15T10:31:00Z",
        "error_message": null
      },
      {
        "instance_id": "instance-2",
        "status": "in_progress",
        "current_version": "v1.2.2",
        "target_version": "v1.2.3",
        "started_at": "2024-01-15T10:30:30Z",
        "completed_at": null,
        "error_message": null
      }
    ],
    "started_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:31:00Z"
  }
}
```

**状态说明**:
- `pending`: 等待执行
- `in_progress`: 执行中
- `completed`: 执行完成
- `failed`: 执行失败
- `cancelled`: 已取消

### 1.3 取消发布任务

**接口描述**: 取消正在执行的发布任务

**请求信息**:
- **URL**: `POST /v1/deploy/cancel/{deploy_id}`
- **Method**: `POST`

**路径参数**:
- `deploy_id`: 发布任务ID

**响应示例**:
```json
{
  "code": 200,
  "message": "deployment cancelled successfully",
  "data": {
    "deploy_id": "deploy-12345",
    "status": "cancelled",
    "cancelled_at": "2024-01-15T10:35:00Z"
  }
}
```

## 2. 版本查询接口

### 2.1 获取服务所有实例的运行版本

**接口描述**: 获取指定服务所有实例的当前运行版本信息

**请求信息**:
- **URL**: `GET /v1/service/{service_name}/instances/versions`
- **Method**: `GET`

**路径参数**:
- `service_name`: 服务名称

**查询参数**:
- `include_stopped`: 是否包含停止的实例，默认false

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "service": "user-service",
    "instances": [
      {
        "instance_id": "instance-1",
        "version": "v1.2.3",
        "status": "running",
        "last_updated": "2024-01-15T10:31:00Z"
      },
      {
        "instance_id": "instance-2",
        "version": "v1.2.2",
        "status": "running",
        "last_updated": "2024-01-15T09:15:00Z"
      }
    ],
    "version_summary": {
      "v1.2.3": 1,
      "v1.2.2": 1
    },
    "total_instances": 2,
    "updated_at": "2024-01-15T10:31:00Z"
  }
}
```

### 2.2 获取服务的实例列表

**接口描述**: 获取指定服务的所有实例详细信息

**请求信息**:
- **URL**: `GET /v1/service/{service_name}/instances`
- **Method**: `GET`

**路径参数**:
- `service_name`: 服务名称

**查询参数**:
- `status`: 实例状态过滤，可选值：running, stopped, error
- `version`: 版本过滤
- `limit`: 返回数量限制，默认100
- `offset`: 偏移量，默认0

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "service": "user-service",
    "instances": [
      {
        "instance_id": "instance-1",
        "host": "192.168.1.10",
        "port": 8080,
        "status": "healthy",
        "version": "v1.2.3",
        "last_heartbeat": "2024-01-15T10:31:00Z",
        "metadata": {
          "region": "us-west-1",
          "zone": "us-west-1a"
        }
      },
      {
        "instance_id": "instance-2",
        "host": "192.168.1.11",
        "port": 8080,
        "status": "healthy",
        "version": "v1.2.2",
        "last_heartbeat": "2024-01-15T10:30:45Z",
        "metadata": {
          "region": "us-west-1",
          "zone": "us-west-1b"
        }
      }
    ],
    "total": 2,
    "limit": 100,
    "offset": 0
  }
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

**接口描述**: 对指定实例执行回滚操作

**请求信息**:
- **URL**: `POST /v1/rollback/instance`
- **Method**: `POST`
- **Content-Type**: `application/json`

**请求参数**:
```json
{
  "instance_id": "string",       // 必填，实例ID
  "target_version": "string",    // 必填，目标版本号
  "package_url": "string",       // 可选，包下载URL（如果不提供，系统会尝试从缓存获取）
  "force": false,                // 可选，是否强制回滚，默认false
  "timeout": 300                 // 可选，超时时间（秒），默认300
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "rollback started successfully",
  "data": {
    "rollback_id": "rollback-67890",
    "instance_id": "instance-1",
    "target_version": "v1.2.2",
    "status": "started",
    "started_at": "2024-01-15T10:55:00Z"
  }
}
```

### 3.2 批量实例回滚

**接口描述**: 对多个实例执行批量回滚操作

**请求信息**:
- **URL**: `POST /v1/rollback/batch`
- **Method**: `POST`
- **Content-Type**: `application/json`

**请求参数**:
```json
{
  "service": "string",           // 必填，服务名称
  "target_version": "string",    // 必填，目标版本号
  "package_url": "string",       // 可选，包下载URL
  "instances": ["string"],       // 必填，需要回滚的实例ID列表
  "force": false,                // 可选，是否强制回滚
  "timeout": 300                 // 可选，超时时间
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "batch rollback started successfully",
  "data": {
    "rollback_id": "rollback-67891",
    "service": "user-service",
    "target_version": "v1.2.2",
    "status": "started",
    "instances": ["instance-1", "instance-2"],
    "total_instances": 2,
    "started_at": "2024-01-15T10:55:00Z"
  }
}
```

### 3.3 查询回滚状态

**接口描述**: 查询指定回滚任务的执行状态

**请求信息**:
- **URL**: `GET /v1/rollback/status/{rollback_id}`
- **Method**: `GET`

**路径参数**:
- `rollback_id`: 回滚任务ID

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "rollback_id": "rollback-67890",
    "service": "user-service",
    "target_version": "v1.2.2",
    "rollback_type": "single",
    "status": "completed",
    "progress": {
      "total": 1,
      "completed": 1,
      "failed": 0,
      "pending": 0
    },
    "instances": [
      {
        "instance_id": "instance-1",
        "status": "completed",
        "current_version": "v1.2.2",
        "target_version": "v1.2.2",
        "started_at": "2024-01-15T10:55:00Z",
        "completed_at": "2024-01-15T11:00:00Z",
        "error_message": null
      }
    ],
    "started_at": "2024-01-15T10:55:00Z",
    "completed_at": "2024-01-15T11:00:00Z"
  }
}
```

### 3.4 取消回滚任务

**接口描述**: 取消正在执行的回滚任务

**请求信息**:
- **URL**: `POST /v1/rollback/cancel/{rollback_id}`
- **Method**: `POST`

**路径参数**:
- `rollback_id`: 回滚任务ID

**响应示例**:
```json
{
  "code": 200,
  "message": "rollback cancelled successfully",
  "data": {
    "rollback_id": "rollback-67890",
    "status": "cancelled",
    "cancelled_at": "2024-01-15T11:05:00Z"
  }
}
```

## 4. 系统管理接口

### 4.1 健康检查

**接口描述**: 检查发布系统健康状态

**请求信息**:
- **URL**: `GET /v1/health`
- **Method**: `GET`

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "status": "healthy",
    "version": "1.0.0",
    "uptime": "72h30m15s",
    "database": "connected",
    "cache": "connected",
    "timestamp": "2024-01-15T11:00:00Z"
  }
}
```

### 4.2 获取系统统计信息

**接口描述**: 获取发布系统的统计信息

**请求信息**:
- **URL**: `GET /v1/stats`
- **Method**: `GET`

**查询参数**:
- `period`: 统计周期，可选值：1h, 24h, 7d, 30d，默认24h

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "period": "24h",
    "deployments": {
      "total": 15,
      "success": 14,
      "failed": 1,
      "success_rate": 93.33
    },
    "rollbacks": {
      "total": 3,
      "success": 3,
      "failed": 0,
      "success_rate": 100.0
    },
    "instances": {
      "total": 50,
      "healthy": 48,
      "unhealthy": 2
    },
    "timestamp": "2024-01-15T11:00:00Z"
  }
}
```

## 5. 使用示例

### 完整发布流程示例

```bash
# 1. 触发发布
curl -X POST http://localhost:8080/v1/deploy/execute \
  -H "Content-Type: application/json" \
  -d '{
    "service": "user-service",
    "version": "v1.2.3",
    "instances": ["instance-1", "instance-2"],
    "package_url": "https://packages.example.com/user-service/v1.2.3.tar.gz",
    "deploy_id": "deploy-12345"
  }'

# 2. 查询发布状态
curl http://localhost:8080/v1/deploy/status/deploy-12345

# 3. 如果发布失败，执行回滚
curl -X POST http://localhost:8080/v1/rollback/batch \
  -H "Content-Type: application/json" \
  -d '{
    "service": "user-service",
    "target_version": "v1.2.2",
    "package_url": "https://packages.example.com/user-service/v1.2.2.tar.gz"
  }'
```
