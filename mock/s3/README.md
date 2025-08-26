# MockS3 - S3兼容的对象存储服务

[![Docker](https://img.shields.io/badge/Docker-Ready-blue?logo=docker)](docker-compose.yml)
[![Go](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go)](go.mod)
[![OpenTelemetry](https://img.shields.io/badge/OpenTelemetry-Enabled-326ce5)](shared/observability)

MockS3 是一个完整的 S3 兼容对象存储服务，专为**混沌工程**和**系统可靠性测试**而设计。它不仅提供标准的对象存储功能，还内置了全面的**错误注入**和**故障模拟**能力。

## ✨ 核心特性

### 🎯 S3兼容API
- 完整的对象 CRUD 操作 (PUT/GET/DELETE/LIST)
- 元数据管理和搜索
- 多存储节点支持
- 异步任务处理

### 💥 错误注入系统
- **CPU峰值**: 模拟高CPU使用率场景
- **内存泄露**: 真实的内存消耗模拟
- **磁盘满载**: 模拟存储空间不足
- **网络风暴**: 大量连接和流量模拟
- **服务宕机**: 完整的服务不可用模拟
- **动态配置**: 通过API实时控制错误注入

### 📊 全栈可观测性
- **OpenTelemetry**: 统一的日志、指标、链路追踪
- **Prometheus + Grafana**: 指标监控和可视化
- **Elasticsearch + Kibana**: 日志分析和搜索
- **分布式追踪**: 跨服务的调用链分析
- **服务发现**: Consul集成

### 微服务组件

| 服务 | 端口 | 职责 |
|-----|------|------|
| **Nginx Gateway** | 8080 | S3 API入口，负载均衡 |
| **Metadata Service** | 8081 | 对象元数据管理 |
| **Storage Service** | 8082 | 文件存储和检索 |
| **Queue Service** | 8083 | 异步任务处理 |
| **Third-Party Service** | 8084 | 外部数据源集成 |
| **Mock Error Service** | 8085 | 错误注入控制中心 |

## 🚀 快速开始

### 前置要求
- Docker 20.10+
- Docker Compose 2.0+

### 部署方式

#### 🚀 一键部署
```bash
# 克隆项目
git clone <repository-url>
cd mock/s3

# 启动所有服务
docker-compose up --build -d

# 检查服务状态
docker-compose ps
```

#### 📦 分模块构建

```bash
# 构建基础设施服务
docker-compose up consul postgres redis -d

# 等待基础设施就绪
docker-compose logs consul | grep "consul: New leader elected"

# 构建观测性服务  
docker-compose up --build otel-collector prometheus -d
docker-compose up --build elasticsearch -d

# 等待ES启动完成
curl -f http://localhost:9200/_cluster/health

# 构建可视化服务
docker-compose up --build grafana kibana -d

# 逐个构建Mock S3服务
docker-compose up --build gateway -d
docker-compose up --build metadata-service -d
docker-compose up --build storage-service -d
docker-compose up --build queue-service -d
docker-compose up --build third-party-service -d
docker-compose up --build mock-error-service -d

# 最终检查所有服务状态
docker-compose ps
```

#### 🔧 资源优化构建
```bash
# 限制并行构建数量，避免内存不足
export COMPOSE_PARALLEL_LIMIT=2

# 逐个构建核心服务
for service in metadata-service storage-service queue-service; do
  echo "Building $service..."
  docker-compose build $service
  docker-compose up $service -d
  sleep 30  # 等待服务启动
done

# 构建剩余服务
docker-compose up --build third-party-service mock-error-service -d
```

#### ⚡ 快速验证构建
```bash
# 只启动核心功能
docker-compose up consul postgres redis metadata-service storage-service -d

# 验证核心功能可用
curl http://localhost:8081/health
curl http://localhost:8082/health
```

### 访问地址
- **S3 API**: http://localhost:8080
- **Consul UI**: http://localhost:8500
- **Grafana监控**: http://localhost:3000 (admin/admin)
- **Kibana日志**: http://localhost:5601
- **Prometheus**: http://localhost:9090

## 📖 使用指南

### S3 API操作

```bash
# 上传对象
curl -X PUT http://localhost:8080/my-bucket/my-object.txt \
  -H "Content-Type: text/plain" \
  -d "Hello MockS3!"

# 下载对象
curl http://localhost:8080/my-bucket/my-object.txt

# 列出对象
curl http://localhost:8081/api/v1/metadata?bucket=my-bucket

# 删除对象
curl -X DELETE http://localhost:8080/my-bucket/my-object.txt
```

### 错误注入

```bash
# 注入CPU峰值异常 (持续2分钟)
curl -X POST http://localhost:8085/api/v1/inject \
  -H "Content-Type: application/json" \
  -d '{
    "service": "storage-service",
    "anomaly_type": "cpu_spike", 
    "target_value": 85.0,
    "duration": "2m"
  }'

# 注入内存泄露 (分配1GB内存)
curl -X POST http://localhost:8085/api/v1/inject \
  -H "Content-Type: application/json" \
  -d '{
    "service": "metadata-service",
    "anomaly_type": "memory_leak",
    "target_value": 1024,
    "duration": "5m"
  }'

# 停止所有异常注入
curl -X POST http://localhost:8085/api/v1/stop-all
```

### 监控和日志

```bash
# 查看服务健康状态
curl http://localhost:8081/health
curl http://localhost:8082/health

# 查看错误注入状态
curl http://localhost:8085/api/v1/status

# 查看队列长度
curl http://localhost:8083/api/v1/queues/status
```

## 🔧 开发指南

### 本地开发环境

```bash
# 只启动基础设施
docker-compose up consul postgres redis elasticsearch -d

# 设置环境变量
export CONSUL_ADDR=localhost:8500
export POSTGRES_HOST=localhost
export REDIS_ADDR=localhost:6379
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318

# 运行单个服务
cd services/metadata
go run cmd/main.go
```

### 项目结构

```
mock/s3/
├── shared/                     # 共享组件
│   ├── interfaces/            # 服务接口定义
│   ├── models/                # 数据模型
│   ├── client/                # HTTP客户端
│   ├── observability/         # 可观测性组件
│   ├── middleware/            # 中间件 (Consul, 错误注入)
│   └── utils/                 # 工具函数
├── services/                  # 微服务实现
│   ├── metadata/              # 元数据服务
│   ├── storage/               # 存储服务
│   ├── queue/                 # 队列服务
│   ├── third-party/           # 第三方服务
│   └── mock-error/            # 错误注入服务
├── gateway/                   # Nginx网关
├── deployments/               # 部署配置
│   ├── consul/               # Consul配置
│   ├── observability/        # 监控配置
│   └── postgres/             # 数据库初始化
└── docker-compose.yml         # 完整堆栈部署
```

### 添加新服务

1. 复制现有服务目录结构
2. 实现对应的接口 (`shared/interfaces/`)
3. 添加服务配置到 `docker-compose.yml`
4. 更新 Consul 服务发现配置

## 📊 监控和可观测性

### 指标监控 (Grafana)
- **系统指标**: CPU、内存、磁盘使用率
- **业务指标**: 请求量、响应时间、错误率
- **服务健康**: 实时健康状态监控

### 日志分析 (Kibana)
- **结构化日志**: JSON格式，支持全文搜索
- **分布式追踪**: trace_id关联的调用链分析
- **错误分析**: 异常日志聚合和分析

### 错误注入监控
- **注入状态**: 实时监控各种异常注入状态
- **资源消耗**: CPU/内存/磁盘/网络的真实消耗
- **影响分析**: 错误注入对系统整体的影响评估

## 🧪 混沌工程实践

### 常见测试场景

```bash
# 场景1: 存储服务高负载
curl -X POST http://localhost:8085/api/v1/inject \
  -d '{"service":"storage-service","anomaly_type":"cpu_spike","duration":"10m"}'

# 场景2: 数据库连接异常
curl -X POST http://localhost:8085/api/v1/inject \
  -d '{"service":"metadata-service","anomaly_type":"machine_down","duration":"30s"}'

# 场景3: 网络拥堵模拟
curl -X POST http://localhost:8085/api/v1/inject \
  -d '{"service":"queue-service","anomaly_type":"network_flood","duration":"5m"}'
```

### 测试指标

- **可用性**: 服务异常时的降级能力
- **性能**: 高负载下的响应时间
- **恢复能力**: 故障恢复的速度
- **数据一致性**: 异常情况下的数据完整性
