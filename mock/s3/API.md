# MockS3 技术文档

> 本文档详细介绍 MockS3 的技术架构、API接口和开发指南

---

## 📋 目录

- [系统架构](#系统架构)
- [API接口文档](#api接口文档)
- [技术栈](#技术栈)
- [项目结构](#项目结构)
- [开发指南](#开发指南)
- [部署配置](#部署配置)
- [监控和可观测性](#监控和可观测性)
- [错误注入机制](#错误注入机制)

---

## 🏗️ 系统架构

### 微服务组件

| 服务 | 端口 | 职责 | 存储 | 依赖 |
|-----|------|------|------|------|
| **Metadata Service** | 8081 | 对象元数据管理 | PostgreSQL | - |
| **Storage Service** | 8082 | 文件存储和检索 | File System | Metadata, Queue, 3rd Party |
| **Queue Service** | 8083 | 异步任务处理 | Redis | Storage |
| **Third-Party Service** | 8084 | 外部数据源集成 | External APIs | - |
| **Mock Error Service** | 8085 | 错误注入控制中心 | File System | - |

---

## 🔌 API接口文档

### Metadata Service API

#### 保存元数据
```http
POST /api/v1/metadata
Content-Type: application/json

{
  "bucket": "test-bucket",
  "key": "test-file.json",
  "size": 1024,
  "content_type": "application/json",
  "md5_hash": "d41d8cd98f00b204e9800998ecf8427e"
}
```

#### 获取元数据
```http
GET /api/v1/metadata/{bucket}/{key}

Response:
{
  "bucket": "test-bucket",
  "key": "test-file.json",
  "size": 1024,
  "content_type": "application/json",
  "md5_hash": "d41d8cd98f00b204e9800998ecf8427e",
  "status": "active",
  "created_at": "2024-01-01T00:00:00Z"
}
```

#### 搜索元数据
```http
GET /api/v1/metadata/search?q=filename&bucket=test-bucket&limit=10

Response:
{
  "query": "filename",
  "objects": [...],
  "total": 42,
  "limit": 10,
  "offset": 0
}
```

#### 统计信息
```http
GET /api/v1/stats

Response:
{
  "total_objects": 1250,
  "total_size_bytes": 52428800,
  "last_updated": "2024-01-01T12:00:00Z"
}
```

### Storage Service API

#### 上传对象
```http
POST /api/v1/objects
Content-Type: application/json

{
  "bucket": "test-bucket",
  "key": "test-file.json",
  "data": "base64-encoded-data",
  "content_type": "application/json",
  "headers": {
    "x-custom-header": "value"
  },
  "tags": {
    "environment": "test"
  }
}
```

#### 下载对象
```http
GET /api/v1/objects/{bucket}/{key}

Response:
Content-Type: application/json
Content-Length: 1024
ETag: "d41d8cd98f00b204e9800998ecf8427e"

{object-content}
```

#### 删除对象
```http
DELETE /api/v1/objects/{bucket}/{key}

Response:
{
  "success": true,
  "message": "Object deleted successfully"
}
```

#### 列出对象
```http
GET /api/v1/objects?bucket=test-bucket&prefix=logs/&max_keys=100

Response:
{
  "bucket": "test-bucket",
  "prefix": "logs/",
  "objects": [
    {
      "key": "logs/app.log",
      "size": 2048,
      "content_type": "text/plain",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "count": 1,
  "is_truncated": false
}
```

### Queue Service API

#### 获取队列统计
```http
GET /api/v1/stats

Response:
{
  "save_queue_length": 5,
  "delete_queue_length": 2,
  "processed_tasks": 1250,
  "failed_tasks": 3,
  "worker_count": 3,
  "last_updated": "2024-01-01T12:00:00Z"
}
```

### Mock Error Service API

#### 创建异常规则
```http
POST /api/v1/metric-anomaly
Content-Type: application/json

{
  "name": "CPU压力测试",
  "service": "storage-service",
  "metric_name": "system_cpu_usage_percent",
  "anomaly_type": "cpu_spike",
  "target_value": 90.0,
  "duration": 120000000000,
  "enabled": true
}
```

#### 检查异常注入状态
```http
POST /api/v1/metric-inject/check
Content-Type: application/json

{
  "service": "storage-service",
  "metric_name": "system_cpu_usage_percent"
}

Response:
{
  "should_inject": true,
  "service": "storage-service",
  "metric_name": "system_cpu_usage_percent",
  "anomaly": {
    "anomaly_type": "cpu_spike",
    "target_value": 90.0,
    "duration": "2m0s",
    "rule_id": "rule-123"
  }
}
```

#### 删除异常规则
```http
DELETE /api/v1/metric-anomaly/{rule_id}

Response:
{
  "success": true,
  "message": "Rule deleted successfully"
}
```

#### 获取错误注入统计
```http
GET /api/v1/stats

Response:
{
  "total_requests": 5420,
  "injected_errors": 127,
  "active_rules": 3,
  "last_updated": "2024-01-01T12:00:00Z"
}
```

### 健康检查API

所有服务都提供健康检查端点：
```http
GET /health

Response:
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "service": "metadata-service"
}
```

---

## 💻 技术栈

### 核心技术
- **编程语言**: Go 1.24
- **Web框架**: Gin
- **容器化**: Docker + Docker Compose
- **服务发现**: Consul
- **可观测性**: OpenTelemetry

### 数据存储
- **关系数据库**: PostgreSQL 15 (元数据)
- **缓存**: Redis 7 (队列和缓存)
- **文件存储**: 本地文件系统 (对象数据)
- **时序数据**: Prometheus (指标)
- **日志存储**: Elasticsearch 8 (日志)

### 监控和可观测性
- **指标监控**: Prometheus + Grafana
- **日志分析**: Elasticsearch + Kibana  
- **链路追踪**: OpenTelemetry
- **服务发现**: Consul

### 依赖管理
详见 `go.mod` 文件中的完整依赖列表。

---

## 📁 项目结构

```
mock/s3/
├── shared/                     # 共享组件
│   ├── interfaces/            # 服务接口定义
│   │   ├── storage.go        # 存储服务接口
│   │   ├── metadata.go       # 元数据服务接口
│   │   ├── queue.go          # 队列服务接口
│   │   └── error_injector.go # 错误注入接口
│   ├── models/               # 数据模型
│   │   ├── object.go        # 对象模型
│   │   ├── metadata.go      # 元数据模型
│   │   ├── task.go          # 任务模型
│   │   ├── error.go         # 错误模型
│   │   └── service.go       # 服务模型
│   ├── client/              # HTTP客户端
│   │   ├── base_client.go   # 基础HTTP客户端
│   │   ├── metadata_client.go
│   │   ├── storage_client.go
│   │   ├── queue_client.go
│   │   └── third_party_client.go
│   ├── observability/       # 可观测性组件
│   │   ├── observability.go # 统一入口
│   │   ├── providers.go     # OpenTelemetry提供者
│   │   ├── logger.go        # 结构化日志
│   │   ├── metrics.go       # 指标收集
│   │   └── middleware.go    # HTTP中间件
│   ├── middleware/          # 中间件
│   │   ├── consul/          # Consul集成
│   │   └── error_injection/ # 错误注入
│   │       ├── error_injection.go      # 主控制器
│   │       ├── cpu_spike_injector.go   # CPU异常
│   │       ├── memory_leak_injector.go # 内存异常
│   │       ├── disk_full_injector.go   # 磁盘异常
│   │       ├── network_flood_injector.go # 网络异常
│   │       └── machine_down_injector.go  # 宕机异常
│   ├── server/              # 服务启动器
│   └── utils/               # 工具函数
├── services/                # 微服务实现
│   ├── metadata/           # 元数据服务
│   │   ├── cmd/main.go    # 服务入口
│   │   ├── internal/
│   │   │   ├── handler/   # HTTP处理器
│   │   │   ├── service/   # 业务逻辑
│   │   │   └── repository/ # 数据访问
│   │   └── config/        # 配置文件
│   ├── storage/           # 存储服务
│   ├── queue/             # 队列服务
│   ├── third-party/       # 第三方服务
│   └── mock-error/        # 错误注入服务
├── deployments/           # 部署配置
│   ├── consul/           # Consul配置
│   ├── observability/    # 监控配置
│   │   ├── grafana/     # Grafana配置
│   │   ├── prometheus.yml
│   │   └── otel-collector-config.yaml
│   └── postgres/        # 数据库初始化
└── docker-compose.yml    # 完整堆栈部署
```

---

## 🛠️ 开发指南

### 环境搭建

#### 1. 安装依赖
```bash
# 安装Go依赖
go mod tidy

# 验证Docker环境
docker --version
docker-compose --version
```

#### 2. 启动开发环境
```bash
# 启动基础设施服务
docker-compose up consul postgres redis -d

# 启动监控服务
docker-compose up prometheus grafana elasticsearch kibana -d

# 本地运行微服务进行开发
cd services/metadata && go run cmd/main.go
cd services/storage && go run cmd/main.go
```

---

## 🚀 部署配置

### Docker配置

#### 1. 基础镜像构建
```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./services/metadata/cmd

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/services/metadata/config ./config
CMD ["./main"]
```

#### 2. 网络配置
```yaml
networks:
  mock-s3-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
```

#### 3. 存储配置
```yaml
volumes:
  postgres-data:
  redis-data: 
  storage-data:
  prometheus-data:
  grafana-data:
  elasticsearch-data:
```

## 📊 监控和可观测性

### OpenTelemetry配置

#### 1. 追踪配置
```yaml
# otel-collector-config.yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:
    timeout: 1s
    send_batch_size: 1024
    
exporters:
  jaeger:
    endpoint: jaeger:14250
    tls:
      insecure: true
  prometheus:
    endpoint: "0.0.0.0:8889"
  elasticsearch:
    endpoints: [http://elasticsearch:9200]
    
service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [jaeger]
    metrics:
      receivers: [otlp]
      processors: [batch]  
      exporters: [prometheus]
    logs:
      receivers: [otlp]
      processors: [batch]
      exporters: [elasticsearch]
```

#### 2. 服务端追踪
```go
// 在每个HTTP处理器中自动生成追踪
func (h *Handler) CreateObject(c *gin.Context) {
    ctx, span := h.tracer.Start(c.Request.Context(), "create_object")
    defer span.End()
    
    // 添加属性
    span.SetAttributes(
        attribute.String("bucket", req.Bucket),
        attribute.String("key", req.Key),
    )
    
    // 调用业务逻辑
    err := h.service.CreateObject(ctx, req)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
    }
}
```

### Prometheus指标

#### 1. 系统指标
- `system_cpu_usage_percent` - CPU使用率
- `system_memory_usage_percent` - 内存使用率
- `system_disk_usage_percent` - 磁盘使用率  
- `system_network_qps` - 网络QPS
- `system_machine_online_status` - 服务在线状态

#### 2. 业务指标
- `http_requests_total` - HTTP请求总数
- `http_request_duration_seconds` - 请求持续时间
- `objects_total` - 对象总数
- `storage_usage_bytes` - 存储使用量
- `queue_length` - 队列长度

#### 3. 错误注入指标
- `anomaly_injection_active` - 异常注入状态
- `anomaly_injection_count` - 异常注入次数
- `resource_consumption_current` - 当前资源消耗

### 日志管理

#### 1. 结构化日志格式
```json
{
  "@timestamp": "2025-08-28T03:34:39.196013946Z",
  "Body": "HTTP request completed",
  "SeverityNumber": 9,
  "TraceId": "36e4c61d27746610192266900c8aa6c7",
  "SpanId": "0525fa6f32bcc3a3",
  "TraceFlags": 1,
  "Attributes": {
    "service": "metadata-service",
    "hostname": "metadata-service",
    "host_address": "172.20.0.31",
    "message": "HTTP request completed",
    "method": "POST",
    "path": "/api/v1/metadata",
    "status": "200",
    "duration": "102.436µs",
    "span_id": "0525fa6f32bcc3a3",
    "trace_id": "36e4c61d27746610192266900c8aa6c7"
  },
  "Resource": {
    "service": {
      "name": "metadata-service",
      "namespace": "mock-s3",
      "version": "1.0.0"
    },
    "deployment": {
      "environment": "development"
    }
  },
  "Scope": {
    "name": "metadata-service",
    "version": ""
  }
}
```

#### 2. 日志级别
- `LevelDebug` - 详细的调试信息
- `LevelInfo` - 正常的操作信息
- `LevelWarn` - 警告但不影响功能  
- `LevelError` - 错误信息需要关注

---

## 💥 错误注入机制

### 支持的异常类型

#### 1. CPU峰值异常 (cpu_spike)
**原理**: 启动多个CPU密集型协程
```go
// 计算所需协程数量
numGoroutines := int(float64(runtime.NumCPU()) * targetCPUPercent / 100.0)

// 启动CPU密集型任务
for i := 0; i < numGoroutines; i++ {
    go func() {
        for {
            select {
            case <-stopChan:
                return
            default:
                // CPU密集型计算
                math.Sqrt(rand.Float64())
            }
        }
    }()
}
```

**参数**:
- `target_value`: 目标CPU使用率 (0-100)
- `duration`: 持续时间 (纳秒)

#### 2. 内存泄露异常 (memory_leak)
**原理**: 真实分配内存并持有引用
```go
func (m *MemoryLeakInjector) allocateMemory(targetMB int64) {
    chunkSize := 1024 * 1024 // 1MB chunks
    
    for m.currentMB < targetMB {
        chunk := make([]byte, chunkSize)
        // 写入数据确保内存真实分配
        for i := range chunk {
            chunk[i] = byte(i % 256)
        }
        m.memoryPool = append(m.memoryPool, chunk)
        m.currentMB++
        
        time.Sleep(100 * time.Millisecond) // 渐进式分配
    }
}
```

**参数**:
- `target_value`: 目标内存使用量 (MB)
- `duration`: 持续时间

#### 3. 磁盘满载异常 (disk_full)
**原理**: 创建大文件占用磁盘空间
```go
func (d *DiskFullInjector) createLargeFile(targetGB int64) error {
    filename := filepath.Join(d.tempDir, fmt.Sprintf("disk-full-%d.tmp", time.Now().Unix()))
    
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()
    
    // 写入指定大小的数据
    data := make([]byte, 1024*1024) // 1MB buffer
    targetBytes := targetGB * 1024 * 1024 * 1024
    
    for written := int64(0); written < targetBytes; written += int64(len(data)) {
        _, err := file.Write(data)
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

**参数**:
- `target_value`: 目标磁盘占用量 (GB)
- `duration`: 持续时间

#### 4. 网络风暴异常 (network_flood)
**原理**: 创建大量网络连接
```go
func (n *NetworkFloodInjector) createConnections(targetConnections int) {
    for i := 0; i < targetConnections; i++ {
        go func() {
            conn, err := net.Dial("tcp", "google.com:80")
            if err != nil {
                return
            }
            
            n.connections = append(n.connections, conn)
            
            // 保持连接活跃
            ticker := time.NewTicker(30 * time.Second)
            defer ticker.Stop()
            
            for {
                select {
                case <-n.stopChan:
                    conn.Close()
                    return
                case <-ticker.C:
                    // 发送keep-alive数据
                    conn.Write([]byte("ping\n"))
                }
            }
        }()
    }
}
```

**参数**:
- `target_value`: 目标连接数
- `duration`: 持续时间

#### 5. 机器宕机异常 (machine_down)
**原理**: 模拟服务挂起或响应延迟
```go
func (m *MachineDownInjector) simulateServiceHang() {
    // 阻塞所有HTTP请求处理
    m.middleware = func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            select {
            case <-m.stopChan:
                next.ServeHTTP(w, r)
            case <-time.After(time.Hour): // 长时间阻塞
                // 请求超时
            }
        })
    }
}
```

**参数**:
- `simulation_type`: 模拟类型 (service_hang, slow_response, connection_refuse)
- `duration`: 持续时间

### 异常注入流程

#### 1. 创建异常规则
```bash
curl -X POST http://localhost:8085/api/v1/metric-anomaly \
  -H "Content-Type: application/json" \
  -d '{
    "name": "内存压力测试",
    "service": "storage-service", 
    "metric_name": "system_memory_usage_percent",
    "anomaly_type": "memory_leak",
    "target_value": 80.0,
    "duration": 300000000000,
    "enabled": true
  }'
```

#### 2. 规则验证和存储
Mock Error Service验证规则有效性并持久化到文件系统。

#### 3. 指标收集时查询
当MetricCollector收集指标时，会查询Mock Error Service：
```go
func (mi *MetricInjector) InjectMetricAnomaly(ctx context.Context, metricName string, originalValue float64) float64 {
    // 检查缓存
    if cached := mi.getFromCache(metricName); cached != nil {
        return mi.applyAnomaly(ctx, cached, originalValue, metricName)
    }
    
    // 查询Mock Error Service
    anomaly := mi.queryMockErrorService(ctx, metricName)
    if anomaly != nil {
        mi.updateCache(metricName, anomaly)
        return mi.applyAnomaly(ctx, anomaly, originalValue, metricName)
    }
    
    return originalValue
}
```

#### 4. 真实资源消耗
根据异常类型启动对应的资源消耗任务：
```go
func (mi *MetricInjector) applyAnomaly(ctx context.Context, anomaly map[string]any, originalValue float64, metricName string) float64 {
    switch anomaly["anomaly_type"].(string) {
    case "cpu_spike":
        if !mi.cpuInjector.IsActive() {
            mi.cpuInjector.StartCPUSpike(ctx, targetValue, duration)
        }
        return targetValue
        
    case "memory_leak":
        if !mi.memoryInjector.IsActive() {
            mi.memoryInjector.StartMemoryLeak(ctx, int64(targetValue), duration)
        }
        return float64(mi.memoryInjector.GetCurrentMemoryMB())
        
    // ... 其他异常类型
    }
}
```

#### 5. 自动清理
异常持续时间结束后，自动清理资源：
```go
func (c *CPUSpikeInjector) StopCPUSpike(ctx context.Context) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    if !c.isActive {
        return
    }
    
    // 停止所有协程
    close(c.stopChan)
    for _, stopChan := range c.goroutines {
        close(stopChan)
    }
    
    c.isActive = false
    c.goroutines = nil
    c.stopChan = make(chan struct{})
}
```

### 缓存机制

为了避免频繁查询Mock Error Service，实现了TTL缓存：

```go
type CachedAnomaly struct {
    Anomaly   map[string]any
    ExpiresAt time.Time
}

func (mi *MetricInjector) updateCache(key string, anomaly map[string]any) {
    mi.cacheMu.Lock()
    defer mi.cacheMu.Unlock()
    
    mi.cache[key] = &CachedAnomaly{
        Anomaly:   anomaly,
        ExpiresAt: time.Now().Add(mi.cacheTTL),
    }
}
```

### 监控异常注入

可以通过以下方式监控异常注入状态：

#### 1. Grafana仪表板
访问 `Mock S3 Services Resource Metrics` 仪表板，观察资源使用率的变化。

#### 2. API查询
```bash
# 查看当前异常注入状态
curl http://localhost:8085/api/v1/stats

# 检查特定服务的异常
curl -X POST http://localhost:8085/api/v1/metric-inject/check \
  -H "Content-Type: application/json" \
  -d '{"service": "storage-service", "metric_name": "system_cpu_usage_percent"}'
```

#### 3. 日志分析
在Kibana中搜索异常注入相关日志：
```
message:"Starting real resource consumption" OR message:"anomaly injection"
```

## API测试

以下提供了完整的API测试示例，演示如何使用curl命令测试所有微服务的接口。

### 1. 系统健康检查

```bash
# 检查所有服务健康状态
curl http://localhost:8081/health  # metadata-service
curl http://localhost:8082/health  # storage-service
curl http://localhost:8083/health  # queue-service
curl http://localhost:8084/health  # third-party-service
curl http://localhost:8085/health  # mock-error-service
```

### 2. Metadata Service API测试

#### 保存元数据
```bash
curl -X POST http://localhost:8081/api/v1/metadata \
  -H "Content-Type: application/json" \
  -d '{
    "bucket": "test-bucket",
    "key": "documents/test-file.pdf",
    "size": 2048,
    "content_type": "application/pdf",
    "md5_hash": "9bb58f26192e4ba00f01e2e7b136bbd8",
    "headers": {
      "x-custom-header": "value"
    },
    "tags": {
      "project": "demo",
      "environment": "test"
    }
  }'
```

#### 获取元数据
```bash
curl http://localhost:8081/api/v1/metadata/test-bucket/documents/test-file.pdf
```

#### 更新元数据
```bash
curl -X PUT http://localhost:8081/api/v1/metadata/test-bucket/documents/test-file.pdf \
  -H "Content-Type: application/json" \
  -d '{
    "bucket": "test-bucket",
    "key": "documents/test-file.pdf",
    "size": 2048,
    "content_type": "application/pdf",
    "md5_hash": "9bb58f26192e4ba00f01e2e7b136bbd8",
    "tags": {
      "project": "demo",
      "environment": "production",
      "version": "v1.1"
    }
  }'
```

#### 列出元数据
```bash
curl "http://localhost:8081/api/v1/metadata?bucket=test-bucket&prefix=documents&limit=10&offset=0"
```

#### 搜索元数据
```bash
curl "http://localhost:8081/api/v1/metadata/search?q=test-file&limit=5"
```

#### 删除元数据
```bash
curl -X DELETE http://localhost:8081/api/v1/metadata/test-bucket/documents/test-file.pdf
```

#### 获取统计信息
```bash
curl http://localhost:8081/api/v1/stats
```

### 3. Storage Service API测试

#### 上传对象
```bash
# 准备测试数据
echo "Hello, MockS3!" > /tmp/test-file.txt

curl -X POST http://localhost:8082/api/v1/objects \
  -H "Content-Type: application/json" \
  -d '{
    "bucket": "storage-test",
    "key": "files/hello.txt",
    "content_type": "text/plain",
    "data": "SGVsbG8sIE1vY2tTMyE=",
    "headers": {
      "cache-control": "max-age=3600"
    },
    "tags": {
      "type": "text-file",
      "source": "api-test"
    }
  }'
```

#### 获取对象
```bash
curl http://localhost:8082/api/v1/objects/storage-test/files/hello.txt
```

#### 更新对象
```bash
curl -X PUT http://localhost:8082/api/v1/objects/storage-test/files/hello.txt \
  -H "Content-Type: application/json" \
  -d '{
    "content_type": "text/plain",
    "data": "SGVsbG8sIFVwZGF0ZWQgTW9ja1MzIQ==",
    "headers": {
      "cache-control": "max-age=7200"
    }
  }'
```

#### 删除对象
```bash
curl -X DELETE http://localhost:8082/api/v1/objects/storage-test/files/hello.txt
```

#### 列出对象
```bash
curl "http://localhost:8082/api/v1/objects?bucket=storage-test&prefix=files&max_keys=10"
```

#### 内部接口测试（仅写入存储）
```bash
curl -X POST http://localhost:8082/api/v1/internal/objects \
  -H "Content-Type: application/json" \
  -d '{
    "bucket": "internal-test",
    "key": "internal-file.txt",
    "content_type": "text/plain",
    "data": "SW50ZXJuYWwgZmlsZSBkYXRh"
  }'
```

#### 内部接口删除
```bash
curl -X DELETE http://localhost:8082/api/v1/internal/objects/internal-test/internal-file.txt
```

#### 获取存储统计
```bash
curl http://localhost:8082/api/v1/stats
```

### 4. Queue Service API测试

#### 删除任务队列操作

##### 入队删除任务
```bash
curl -X POST http://localhost:8083/api/v1/delete-tasks \
  -H "Content-Type: application/json" \
  -d '{
    "object_key": "storage-test/files/to-delete.txt"
  }'
```

##### 出队删除任务
```bash
curl http://localhost:8083/api/v1/delete-tasks/dequeue
```

##### 更新删除任务状态
```bash
# 获取任务ID后更新状态
curl -X PUT http://localhost:8083/api/v1/delete-tasks/del_12345678-1234-5678-9abc-123456789012/status \
  -H "Content-Type: application/json" \
  -d '{
    "status": "completed"
  }'

# 标记任务失败
curl -X PUT http://localhost:8083/api/v1/delete-tasks/del_12345678-1234-5678-9abc-123456789012/status \
  -H "Content-Type: application/json" \
  -d '{
    "status": "failed",
    "error": "File not found in storage"
  }'
```

#### 保存任务队列操作

##### 入队保存任务
```bash
curl -X POST http://localhost:8083/api/v1/save-tasks \
  -H "Content-Type: application/json" \
  -d '{
    "object_key": "third-party/data/export.csv",
    "object": {
      "bucket": "backups",
      "key": "third-party/data/export.csv",
      "size": 4096,
      "content_type": "text/csv",
      "data": "bmFtZSxhZ2UsY2l0eQpKb2huLDMwLE5ldyBZb3Jr",
      "headers": {
        "x-source": "third-party-api"
      }
    }
  }'
```

##### 出队保存任务
```bash
curl http://localhost:8083/api/v1/save-tasks/dequeue
```

##### 更新保存任务状态
```bash
curl -X PUT http://localhost:8083/api/v1/save-tasks/save_87654321-4321-8765-dcba-210987654321/status \
  -H "Content-Type: application/json" \
  -d '{
    "status": "completed"
  }'
```

#### 获取队列统计
```bash
curl http://localhost:8083/api/v1/stats
```

### 5. Third-Party Service API测试

#### 获取第三方对象
```bash
curl http://localhost:8084/api/v1/objects/external-bucket/data/report.json
```

#### 获取第三方服务统计
```bash
curl http://localhost:8084/api/v1/stats
```

### 6. Mock Error Service API测试

#### 创建CPU峰值异常规则
```bash
curl -X POST http://localhost:8085/api/v1/metric-anomaly \
  -H "Content-Type: application/json" \
  -d '{
    "id": "cpu_spike_test_001",
    "name": "存储服务CPU异常测试",
    "service": "storage-service",
    "metric_name": "system_cpu_usage_percent",
    "anomaly_type": "cpu_spike",
    "enabled": true,
    "target_value": 95.0,
    "duration": 300000000000,
    "max_triggers": 5
  }'
```

#### 创建内存泄露异常规则
```bash
curl -X POST http://localhost:8085/api/v1/metric-anomaly \
  -H "Content-Type: application/json" \
  -d '{
    "id": "memory_leak_test_001",
    "name": "元数据服务内存异常测试",
    "service": "metadata-service",
    "metric_name": "system_memory_usage_percent",
    "anomaly_type": "memory_leak",
    "enabled": true,
    "target_value": 92.5,
    "duration": 600000000000,
    "max_triggers": 3
  }'
```

#### 创建磁盘满异常规则
```bash
curl -X POST http://localhost:8085/api/v1/metric-anomaly \
  -H "Content-Type: application/json" \
  -d '{
    "id": "disk_full_test_001",
    "name": "队列服务磁盘异常测试",
    "service": "queue-service",
    "metric_name": "system_disk_usage_percent",
    "anomaly_type": "disk_full",
    "enabled": true,
    "target_value": 98.0,
    "duration": 180000000000,
    "max_triggers": 2
  }'
```

#### 创建网络洪泛异常规则
```bash
curl -X POST http://localhost:8085/api/v1/metric-anomaly \
  -H "Content-Type: application/json" \
  -d '{
    "id": "network_flood_test_001",
    "name": "第三方服务网络异常测试",
    "service": "third-party-service",
    "metric_name": "system_network_qps",
    "anomaly_type": "network_flood",
    "enabled": true,
    "target_value": 10000.0,
    "duration": 120000000000,
    "max_triggers": 1
  }'
```

#### 创建机器宕机异常规则
```bash
curl -X POST http://localhost:8085/api/v1/metric-anomaly \
  -H "Content-Type: application/json" \
  -d '{
    "id": "machine_down_test_001",
    "name": "存储服务宕机测试",
    "service": "storage-service",
    "metric_name": "system_machine_online_status",
    "anomaly_type": "machine_down",
    "enabled": true,
    "target_value": 80.0,
    "duration": 60000000000,
    "max_triggers": 1
  }'
```

#### 检查异常注入状态
```bash
curl -X POST http://localhost:8085/api/v1/metric-inject/check \
  -H "Content-Type: application/json" \
  -d '{
    "service": "storage-service",
    "metric_name": "system_cpu_usage_percent"
  }'
```

#### 删除异常规则
```bash
curl -X DELETE http://localhost:8085/api/v1/metric-anomaly/cpu_spike_test_001
```

#### 获取异常注入统计
```bash
curl http://localhost:8085/api/v1/stats
```

### 7. 完整工作流程测试

以下是一个完整的测试流程，演示系统各组件如何协作：

#### 步骤1：上传文件到存储服务
```bash
curl -X POST http://localhost:8082/api/v1/objects \
  -H "Content-Type: application/json" \
  -d '{
    "bucket": "test-bucket",
    "key": "test-file.json",
    "data": "'$(cat test-file.json | base64)'",
    "content_type": "application/json"
  }' | jq .
```

#### 步骤2：验证上传结果
```bash
curl "http://localhost:8082/api/v1/objects?bucket=test-bucket" | jq .
```

#### 步骤3. 下载并验证内容
```bash
curl http://localhost:8082/api/v1/objects/test-bucket/test-file.json
```

#### 步骤4. 更新元数据
```bash
curl -X POST http://localhost:8081/api/v1/metadata \
  -H "Content-Type: application/json" \
  -d "{
    \"bucket\": \"test-bucket\",
    \"key\": \"test-file.json\",
    \"size\": $(wc -c < test-file.json),
    \"content_type\": \"application/json\"
  }" | jq .
```

#### 步骤5. 查看统计信息
```bash
echo "=== Storage Stats ===" && curl -s http://localhost:8082/api/v1/stats | jq .
echo "=== Metadata Stats ===" && curl -s http://localhost:8081/api/v1/stats | jq .
```

#### 步骤6. 删除对象
```bash
curl -X DELETE http://localhost:8082/api/v1/objects/test-bucket/test-file.json
```

#### 步骤7. 验证删除结果
```bash
curl "http://localhost:8082/api/v1/objects?bucket=test-bucket" | jq .
```

### 8. 监控与可观测性测试

#### 查看Prometheus指标
```bash
curl http://localhost:9090/api/v1/query?query=system_cpu_usage_percent
curl http://localhost:9090/api/v1/query?query=system_memory_usage_percent
```

#### 查看服务注册状态
```bash
curl http://localhost:8500/v1/catalog/services
curl http://localhost:8500/v1/health/service/storage-service
```

### 9. 错误场景测试

#### 测试无效请求
```bash
# 缺少必需字段
curl -X POST http://localhost:8082/api/v1/objects \
  -H "Content-Type: application/json" \
  -d '{
    "bucket": "test-bucket"
  }'
# 应返回400 Bad Request

# 访问不存在的对象
curl http://localhost:8082/api/v1/objects/non-existent/file.txt
# 应返回404 Not Found
```

#### 测试异常注入效果
```bash
# 创建CPU异常后观察指标变化
curl -X POST http://localhost:8085/api/v1/metric-anomaly \
  -H "Content-Type: application/json" \
  -d '{
    "id": "test_cpu_spike",
    "service": "storage-service",
    "metric_name": "system_cpu_usage_percent",
    "anomaly_type": "cpu_spike",
    "enabled": true,
    "target_value": 90.0,
    "duration": 60000000000
  }'

# 等待30秒后查看CPU使用率
sleep 30
curl http://localhost:9090/api/v1/query?query=system_cpu_usage_percent{service=\"storage-service\"}
```
