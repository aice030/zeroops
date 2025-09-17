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

## 3. API接口设计

### 3.1 发布相关接口

#### 3.1.1 触发发布
```http
POST /v1/deploy/execute
Content-Type: application/json

{
  "service": "user-service",
  "version": "v1.2.3",
  "instances": ["instance-1", "instance-2", "instance-3"],
  "package_url": "https://packages.example.com/user-service/v1.2.3.tar.gz",
  "deploy_id": "deploy-12345"
}
```

**响应：**
```json
{
  "deploy_id": "deploy-12345",
  "status": "started",
  "message": "deployment started successfully",
  "started_at": "2024-01-15T10:30:00Z"
}
```

#### 3.1.2 查询发布状态
```http
GET /v1/deploy/status/{deploy_id}
```

**响应：**
```json
{
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
      "version": "v1.2.3",
      "updated_at": "2024-01-15T10:31:00Z"
    },
    {
      "instance_id": "instance-2",
      "status": "in_progress",
      "version": "v1.2.2",
      "updated_at": "2024-01-15T10:30:30Z"
    }
  ],
  "started_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:31:00Z"
}
```

### 3.2 版本查询接口

#### 3.2.1 获取服务所有实例的运行版本
```http
GET /v1/service/{service_name}/instances/versions
```

**响应：**
```json
{
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
  }
}
```

#### 3.2.2 获取服务的实例列表
```http
GET /v1/service/{service_name}/instances
```

**响应：**
```json
{
  "service": "user-service",
  "instances": [
    {
      "instance_id": "instance-1",
      "host": "192.168.1.10",
      "port": 8080,
      "status": "healthy",
      "version": "v1.2.3"
    },
    {
      "instance_id": "instance-2",
      "host": "192.168.1.11",
      "port": 8080,
      "status": "healthy",
      "version": "v1.2.2"
    }
  ],
  "total": 2
}
```

### 3.3 回滚相关接口

#### 3.3.1 单实例回滚
```http
POST /v1/rollback/instance
Content-Type: application/json

{
  "instance_id": "instance-1",
  "target_version": "v1.2.2",
  "package_url": "https://packages.example.com/user-service/v1.2.2.tar.gz"
}
```

**响应：**
```json
{
  "rollback_id": "rollback-67890",
  "instance_id": "instance-1",
  "target_version": "v1.2.2",
  "status": "started",
  "message": "rollback started successfully"
}
```

#### 3.3.2 批量实例回滚
```http
POST /v1/rollback/batch
Content-Type: application/json

{
  "service": "user-service",
  "target_version": "v1.2.2",
  "package_url": "https://packages.example.com/user-service/v1.2.2.tar.gz",
  "instances": ["instance-1", "instance-2"]
}
```

**响应：**
```json
{
  "rollback_id": "rollback-67891",
  "service": "user-service",
  "target_version": "v1.2.2",
  "status": "started",
  "instances": ["instance-1", "instance-2"],
  "message": "batch rollback started successfully"
}
```

#### 3.3.3 查询回滚状态
```http
GET /v1/rollback/status/{rollback_id}
```

**响应：**
```json
{
  "rollback_id": "rollback-67890",
  "service": "user-service",
  "target_version": "v1.2.2",
  "status": "completed",
  "instances": [
    {
      "instance_id": "instance-1",
      "status": "completed",
      "version": "v1.2.2",
      "updated_at": "2024-01-15T11:00:00Z"
    }
  ],
  "started_at": "2024-01-15T10:55:00Z",
  "completed_at": "2024-01-15T11:00:00Z"
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
