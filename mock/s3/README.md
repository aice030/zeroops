# MockS3

[![Docker](https://img.shields.io/badge/Docker-Ready-blue?logo=docker)](docker-compose.yml)
[![Go](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go)](go.mod)
[![OpenTelemetry](https://img.shields.io/badge/OpenTelemetry-Enabled-326ce5)](shared/observability)

**MockS3** 是一个专为**智能运维**打造的**故障模拟平台**。通过真实的资源消耗模拟生产环境故障场景，为时序异常检测提供高质量的训练数据，解决传统简单Mock数据质量差、关联性缺失的根本问题。

---

## 🎯 设计目标

### 核心理念
构建一个**高保真度**的分布式系统故障模拟平台，专注于：

- **真实性** - 通过实际资源消耗而非简单数值模拟来产生故障
- **可观测性** - 全链路监控和追踪，确保每个故障都能被准确捕获
- **可控性** - 精确控制故障的类型、强度、持续时间和影响范围
- **实用性** - 为AI异常检测算法提供高质量的训练和测试数据

### 技术目标
- **微服务架构** - 模拟真实生产环境的服务间依赖关系
- **故障隔离** - 支持单服务和跨服务的故障注入场景
- **指标完整性** - 覆盖系统层面（CPU、内存、网络、磁盘）和业务层面的指标

---

## 💡 核心功能

### 🔬 **故障注入引擎**
5种真实资源消耗的故障场景：

| 故障类型 | 实现方式 | 监控指标 |
|---------|---------|----------|
| **CPU峰值** | 真实CPU密集计算 | CPU使用率、负载、响应时间 |
| **内存泄露** | 实际分配内存不释放 | 内存使用率、GC频率、OOM事件 |
| **网络风暴** | 大量并发连接 | 网络带宽、连接数、超时率 |
| **服务宕机** | 完整服务停止响应 | 服务健康状态、请求成功率 |

### 🎯 **S3兼容存储**
- 完整的对象CRUD操作和标准S3 API兼容性
- 多存储节点支持和元数据管理
- 支持分布式存储场景的故障模拟

### 📊 **一体化监控**
基于OpenTelemetry的完整可观测性栈：
- **Prometheus + Grafana** - 实时指标监控和可视化仪表板
- **Consul** - 服务发现和健康检查
- **分布式追踪** - 完整的调用链路分析

### 🧩 **核心组件**

#### 微服务组件
| 服务名称 | 端口 | 主要职责 |
|---------|------|----------|
| **Metadata Service** | 8081 | 对象元数据管理、搜索、统计 |
| **Storage Service** | 8082 | 文件存储、检索、多节点管理 |
| **Queue Service** | 8083 | 异步任务队列、工作器管理 |
| **Third-Party Service** | 8084 | 外部数据源集成、API适配 |
| **Mock Error Service** | 8085 | 故障注入规则、资源消耗控制 |

#### 基础设施组件
| 组件名称 | 端口 | 主要用途 |
|---------|------|----------|
| **Consul** | 8500 | 服务发现、配置管理、健康检查 |
| **PostgreSQL** | 5432 | 元数据持久化存储 |
| **Redis** | 6379 | 任务队列、缓存 |
| **Prometheus** | 9090 | 指标数据收集和存储 |
| **Grafana** | 3000 | 指标可视化仪表板 |
| **Elasticsearch** | 9200 | 日志和追踪数据存储 |
| **Kibana** | 5601 | 日志分析和查询界面 |
| **OpenTelemetry Collector** | 4317/4318 | 遥测数据收集和转发 |

#### 网络架构
```
Docker网络：172.20.0.0/16 (支持动态多实例扩容)
├─ 基础设施层：固定IP (172.20.0.10-29)
│  ├─ consul: 172.20.0.10
│  ├─ postgres: 172.20.0.11  
│  ├─ redis: 172.20.0.12
│  └─ 监控组件: 172.20.0.20-24
└─ 业务服务层：动态分配IP (支持多实例)
   ├─ metadata-service: 可扩容到多个实例
   ├─ storage-service: 可扩容到多个实例
   ├─ queue-service: 可扩容到多个实例
   ├─ third-party-service: 可扩容到多个实例
   └─ mock-error-service: 可扩容到多个实例

服务发现：
• 每个服务实例使用UUID生成唯一ServiceID
• Consul自动负载均衡到健康实例
• 端口范围映射支持多实例访问

技术实现：
• ServiceID格式: {service-name}-{uuid}
• 支持动态扩缩容: docker-compose up -d --scale service=N
• 无状态设计: 实例间无依赖，可随意增减
```

---

## 🏢 适用场景

### 🧪 **混沌工程实践**
- **系统韧性验证** - 测试分布式系统在故障下的自愈和降级能力
- **性能基准测试** - 确定系统性能边界和识别潜在瓶颈
- **监控有效性验证** - 确保告警系统能及时准确地发现各类异常

### 🤖 **AI运维场景**
- **异常检测训练** - 为时序异常检测模型提供标注的真实数据
- **根因分析验证** - 模拟复杂故障传播链路，验证分析算法
- **预测模型测试** - 测试故障预测算法在真实场景下的准确性

---

## 🚀 快速开始

### 第一步：环境准备
```bash
# 确保Docker环境可用
docker --version && docker-compose --version
```

### 第二步：启动服务栈

#### 单实例模式
```bash
# 启动完整服务栈（每个服务1个实例）
docker-compose up --build -d

# 等待服务就绪
docker-compose ps
```

#### 多实例模式

**一次性构建**
```bash
# 启动多实例服务栈
docker-compose up --build -d \
  --scale metadata-service=3 \
  --scale storage-service=2 \
  --scale queue-service=2 \
  --scale third-party-service=2 \
  --scale mock-error-service=1
```

**分批构建**
```bash
# 第一步：启动基础设施服务
docker-compose up -d consul postgres redis elasticsearch prometheus grafana kibana otel-collector

# 第二步：分别构建各个服务镜像
docker-compose build metadata-service
docker-compose build storage-service  
docker-compose build queue-service
docker-compose build third-party-service
docker-compose build mock-error-service

# 第三步：分批启动业务服务
docker-compose up -d --scale metadata-service=3 metadata-service
docker-compose up -d --scale storage-service=2 storage-service
docker-compose up -d --scale queue-service=2 queue-service
docker-compose up -d --scale third-party-service=2 third-party-service
docker-compose up -d --scale mock-error-service=1 mock-error-service

# 验证所有实例运行状态
docker-compose ps

# 查看Consul服务发现状态
curl -s "http://localhost:8500/v1/catalog/services" | jq .
```

### 第三步：故障注入体验
```bash
# 创建CPU峰值异常 - 持续2分钟
curl -X POST http://localhost:8085/api/v1/metric-anomaly \
  -H "Content-Type: application/json" \
  -d '{
    "name": "CPU压力测试",
    "service": "storage-service",
    "metric_name": "system_cpu_usage_percent",
    "anomaly_type": "cpu_spike",
    "target_value": 85.0,
    "duration": 120000000000,
    "enabled": true
  }'

# 观察Grafana中CPU使用率的变化
# 访问: http://localhost:3000 → Mock S3 Services Resource Metrics
```

---

## 📖 核心业务流程

### 🔄 **对象存储业务流程**

#### 上传对象流程
```
客户端 -> 存储服务 -> 存储节点 (并行写入所有节点)
        |         -> 元数据服务 (保存元数据)
        |         -> 队列服务 (失败时异步清理)
        |
        <- 返回结果
```

**核心逻辑**：
- **并行写入** - 文件同时写入所有存储节点确保冗余
- **事务保证** - 元数据保存失败时自动清理已写入文件
- **MD5校验** - 自动计算并验证文件完整性

#### 下载对象流程
```
客户端 -> 存储服务 -> 元数据服务 (查询对象元数据)
        |         -> 存储节点 (从任一可用节点读取)
        |         -> 第三方服务 (本地失败时备份获取)
        |         -> 队列服务 (第三方数据异步保存)
        |
        <- 返回文件数据
```

**核心逻辑**：
- **元数据优先** - 通过元数据服务确认对象存在
- **节点容错** - 从任一可用存储节点读取数据
- **第三方备份** - 本地失败时从第三方服务获取
- **异步同步** - 第三方数据异步保存到本地存储

#### 删除对象流程
```
客户端 -> 存储服务 -> 元数据服务 (立即删除元数据)
        |         -> 队列服务 (异步删除任务)
        |
        <- 立即返回成功
        
队列工作器 -> 存储节点 (从所有节点删除文件)
           -> 队列服务 (更新任务状态)
```

**核心逻辑**：
- **即时响应** - 元数据删除后立即返回，文件异步清理
- **任务队列** - 通过Redis队列管理删除任务
- **全节点清理** - 从所有存储节点删除文件
- **状态跟踪** - 完整的任务状态管理

### 🎯 **故障注入工作流程**

#### 故障生命周期
```
故障规则创建 -> 规则验证 -> 规则存储
             |
             v
指标收集查询 -> 缓存检查 -> 真实资源消耗
             |
             v
指标异常产生 -> 监控告警 -> 自动清理
```

**实现机制**：
- **规则管理** - Mock Error Service管理异常规则
- **查询缓存** - 指标收集时查询异常配置（支持TTL缓存）
- **真实消耗** - 启动实际的CPU/内存/磁盘/网络资源消耗
- **自动清理** - 超时后自动释放资源和停止异常
