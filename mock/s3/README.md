# MockS3 - 企业级混沌工程平台

[![Docker](https://img.shields.io/badge/Docker-Ready-blue?logo=docker)](docker-compose.yml)
[![Go](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go)](go.mod)
[![OpenTelemetry](https://img.shields.io/badge/OpenTelemetry-Enabled-326ce5)](shared/observability)

**MockS3** 是一个专为智能运维打造的**混沌工程和系统可靠性测试平台**。基于S3兼容API，提供完整的对象存储功能，同时内置强大的故障注入能力，帮助团队构建更可靠的分布式系统。

---

## 🎯 产品价值

### 为什么实现 MockS3？

#### 🔬 **真实的故障模拟**
不是简单的数值模拟，而是**真实消耗系统资源**来模拟故障场景：
- CPU峰值导致的服务响应缓慢
- 内存泄露引起的系统不稳定  
- 磁盘满载造成的写入失败
- 网络拥塞影响的服务通信
- 完整的服务宕机场景

#### 📊 **完善的监控体系**
基于OpenTelemetry构建的完整可观测性方案：
- **实时指标监控** - Prometheus + Grafana
- **日志分析** - Elasticsearch + Kibana  
- **链路追踪** - 分布式调用链分析
- **服务健康** - 自动服务发现和健康检查

#### ⚡ **开箱即用**
一键部署，无需复杂配置：
- Docker Compose 全栈部署
- 预配置的监控仪表板
- 支持本地和云环境

---

## 🏢 适用场景

### 🧪 **混沌工程实践**
- **可用性测试**: 验证系统在异常情况下的降级能力
- **性能边界**: 探索系统在高负载下的性能极限
- **恢复能力**: 测试故障恢复的速度和数据一致性
- **容错机制**: 验证系统的容错设计是否有效

---

## 🚀 快速开始

### 第一步：环境准备
```bash
# 确保Docker环境可用
docker --version && docker-compose --version
```

### 第二步：一键启动
```bash

# 启动完整服务栈
docker-compose up --build -d

# 等待服务就绪
docker-compose ps
```

### 第三步：访问控制台
- **监控仪表板**: http://localhost:3000 (Grafana - admin/admin)
- **服务发现**: http://localhost:8500 (Consul)
- **日志分析**: http://localhost:5601 (Kibana)
- **指标查询**: http://localhost:9090 (Prometheus)

### 第四步：故障注入体验
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

## 📋 核心功能

### 🎯 **S3兼容存储**
- 完整的对象CRUD操作
- 标准的S3 API兼容性
- 多存储节点支持
- 元数据管理和搜索

### 💥 **5种故障场景**
| 故障类型 | 描述 | 用途 |
|---------|------|------|
| **CPU峰值** | 真实的CPU密集计算 | 测试高负载下的服务响应 |
| **内存泄露** | 实际分配内存不释放 | 验证内存监控和自动重启 |
| **磁盘满载** | 创建大文件占用磁盘 | 测试存储空间不足的处理 |
| **网络风暴** | 大量并发连接 | 验证网络拥塞下的服务稳定性 |
| **服务宕机** | 完整服务不响应 | 测试服务发现和故障转移 |

### 📊 **实时监控**
- **服务指标**: 每个微服务的独立资源监控
- **业务指标**: 存储使用量、请求成功率、队列长度
- **异常状态**: 错误注入的实时状态和影响
- **系统健康**: 自动健康检查和告警
