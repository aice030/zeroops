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
curl -f http://100.100.57.39:9200/_cluster/health

# 构建可视化服务
docker-compose up --build grafana kibana -d

# 逐个构建Mock S3服务
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
curl http://100.100.57.39:8081/health
curl http://100.100.57.39:8082/health
```

### 访问地址
- **Consul UI**: http://100.100.57.39:8500
- **Grafana监控**: http://100.100.57.39:3000 (admin/admin)
- **Kibana日志**: http://100.100.57.39:5601
- **Prometheus**: http://100.100.57.39:9090

## 📖 使用指南

### S3 API操作

### 监控和日志

#### 服务健康检查
```bash
# 查看所有服务健康状态
curl http://100.100.57.39:8081/health  # Metadata Service
curl http://100.100.57.39:8082/health  # Storage Service
curl http://100.100.57.39:8083/health  # Queue Service
curl http://100.100.57.39:8084/health  # Third-Party Service
curl http://100.100.57.39:8085/health  # Mock Error Service

# 查看服务注册状态 (Consul)
docker exec mock-s3-consul consul catalog services -tags
```

#### 业务统计监控
```bash
# Storage Service统计
curl http://100.100.57.39:8082/api/v1/stats
# 返回: 存储节点状态、总存储空间

# Metadata Service统计
curl http://100.100.57.39:8081/api/v1/stats
# 返回: 对象总数、总大小、最后更新时间

# Queue Service统计
curl http://100.100.57.39:8083/api/v1/stats
# 返回: 保存队列、删除队列长度

# Third-Party Service统计
curl http://100.100.57.39:8084/api/v1/stats
# 返回: 数据源状态、成功率配置

# Mock Error Service统计
curl http://100.100.57.39:8085/api/v1/stats
# 返回: 总请求数、错误注入次数
```

#### 指标监控 (Prometheus)
```bash
# 查看系统状态
curl "http://100.100.57.39:9090/api/v1/query?query=up"

# 查看HTTP请求指标
curl "http://100.100.57.39:9090/api/v1/query?query=prometheus_http_requests_total"

# 访问Prometheus UI: http://100.100.57.39:9090
```

#### 日志查看 (Elasticsearch + Kibana)
```bash
# 查看日志总数
curl "http://100.100.57.39:9200/mock-s3-logs/_count"

# 查看最新日志
curl -s "http://100.100.57.39:9200/mock-s3-logs/_search?sort=@timestamp:desc&size=5" | \
  jq -r '.hits.hits[]._source | [."@timestamp", .Body] | @tsv'

# 查看成功操作日志
curl -s "http://100.100.57.39:9200/mock-s3-logs/_search?q=Body:*object*&size=5"

# 访问Kibana UI: http://100.100.57.39:5601
```

#### 链路追踪 (OpenTelemetry)
```bash
# 查看Trace数量
curl "http://100.100.57.39:9200/mock-s3-traces/_count"

# 检查OTEL Collector状态
curl "http://100.100.57.39:13133/"

# 查看链路追踪样例
curl -s "http://100.100.57.39:9200/mock-s3-traces/_search?size=2" | \
  jq -r '.hits.hits[]._source | [."@timestamp", .TraceId[0:8], .SpanId[0:8]] | @tsv'
```

## 完整测试示例

### 端到端S3操作测试
```bash

# 1. 上传对象
curl -X POST http://100.100.57.39:8082/api/v1/objects \
  -H "Content-Type: application/json" \
  -d '{
    "bucket": "test-bucket",
    "key": "test-file.json",
    "data": "'$(cat test-file.json | base64)'",
    "content_type": "application/json"
  }' | jq .

# 2. 验证上传结果
curl "http://100.100.57.39:8082/api/v1/objects?bucket=test-bucket" | jq .

# 3. 下载并验证内容
curl http://100.100.57.39:8082/api/v1/objects/test-bucket/test-file.json

# 4. 更新元数据
curl -X POST http://100.100.57.39:8081/api/v1/metadata \
  -H "Content-Type: application/json" \
  -d "{
    \"bucket\": \"test-bucket\",
    \"key\": \"test-file.json\",
    \"size\": $(wc -c < test-file.json),
    \"content_type\": \"application/json\"
  }" | jq .

# 5. 查看统计信息
echo "=== Storage Stats ===" && curl -s http://100.100.57.39:8082/api/v1/stats | jq .
echo "=== Metadata Stats ===" && curl -s http://100.100.57.39:8081/api/v1/stats | jq .

# 6. 删除对象
curl -X DELETE http://100.100.57.39:8082/api/v1/objects/test-bucket/test-file.json

# 7. 验证删除结果
curl "http://100.100.57.39:8082/api/v1/objects?bucket=test-bucket" | jq .
```

### 队列任务测试
```bash
# 创建保存任务
curl -X POST http://100.100.57.39:8083/api/v1/save-tasks \
  -H "Content-Type: application/json" \
  -d '{
    "bucket": "test-bucket",
    "key": "queue-test.txt",
    "content_type": "text/plain",
    "size": 100
  }'

# 查看队列统计
curl http://100.100.57.39:8083/api/v1/stats | jq .
```
## 项目结构

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
curl -X POST http://100.100.57.39:8085/api/v1/inject \
  -d '{"service":"storage-service","anomaly_type":"cpu_spike","duration":"1m"}'

# 场景2: 数据库连接异常
curl -X POST http://100.100.57.39:8085/api/v1/inject \
  -d '{"service":"metadata-service","anomaly_type":"machine_down","duration":"30s"}'

# 场景3: 网络拥堵模拟
curl -X POST http://100.100.57.39:8085/api/v1/inject \
  -d '{"service":"queue-service","anomaly_type":"network_flood","duration":"5m"}'
```

### 测试指标

- **可用性**: 服务异常时的降级能力
- **性能**: 高负载下的响应时间
- **恢复能力**: 故障恢复的速度
- **数据一致性**: 异常情况下的数据完整性
