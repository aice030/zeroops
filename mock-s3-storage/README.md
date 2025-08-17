## 选用技术

- 容器化: Docker
- 服务通信: Http —> gRPC（备选）
- 服务发现: Consul
- API网关: Nginx
- 数据库: PostgreSQL + Redis缓存
- 配置管理: Consul KV
- 消息队列: Redis streams —> Kafka（和缓存都是Redis，但要分成两个实例）
- 监控: Prometheus
- 日志: ElasticSearch + PostgreSQL 
- 链路: ElasticSearch —> Jaeger+ElasticSearch / Grafna Tempo+Kodo（备选）

## 目录结构

```
mock-s3-storage/
├── README.md
├── docker-compose.yml                    # 本地开发环境
├── docker-compose.prod.yml              # 生产环境
├── Makefile                             # 全局构建和管理脚本
├── .env                                 # 环境变量配置
├── .gitignore
│
├── services/                            # 微服务目录
│   ├── gateway/                         # API网关服务 (Nginx)
│   │   ├── nginx/
│   │   │   ├── nginx.conf              # Nginx主配置
│   │   │   ├── conf.d/                 # 服务配置目录
│   │   │   │   ├── upstream.conf       # 上游服务配置
│   │   │   │   ├── s3-api.conf        # S3 API路由
│   │   │   │   ├── admin-api.conf     # 管理API路由
│   │   │   │   └── error-injection.conf # 错误注入路由
│   │   │   └── templates/              # 配置模板
│   │   ├── Dockerfile
│   │   └── README.md
│   │
│   ├── s3-api/                         # S3兼容API服务(业务编排层)
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── handler/
│   │   │   │   ├── s3.go              # S3标准API处理器
│   │   │   │   ├── bucket.go          # Bucket操作
│   │   │   │   ├── object.go          # Object操作
│   │   │   │   └── health.go          # 健康检查
│   │   │   ├── service/
│   │   │   │   ├── orchestrator.go    # 业务流程编排
│   │   │   │   ├── object.go          # 对象操作编排逻辑
│   │   │   │   ├── bucket.go          # 存储桶操作编排
│   │   │   │   ├── validation.go      # 输入验证
│   │   │   │   └── compensation.go    # 补偿事务处理
│   │   │   ├── client/                # 下游服务客户端
│   │   │   │   ├── metadata.go        # 元数据服务客户端
│   │   │   │   ├── storage.go         # 存储服务客户端
│   │   │   │   ├── async.go           # 异步服务客户端
│   │   │   │   └── error_injection.go # 错误注入服务客户端
│   │   │   ├── middleware/
│   │   │   │   ├── auth.go            # S3认证中间件
│   │   │   │   ├── cors.go            # CORS处理
│   │   │   │   ├── circuit_breaker.go # 熔断器
│   │   │   │   ├── retry.go           # 重试机制
│   │   │   │   ├── tracing.go         # 链路追踪
│   │   │   │   ├── metrics.go         # 监控指标
│   │   │   │   └── logging.go         # 日志记录
│   │   │   ├── model/
│   │   │   │   ├── s3.go             # S3数据模型
│   │   │   │   ├── request.go         # 请求模型
│   │   │   │   └── response.go        # 响应模型
│   │   │   └── config/
│   │   │       └── config.go
│   │   ├── api/
│   │   │   └── swagger/
│   │   │       ├── s3-api.yaml       # S3 API规范
│   │   │       └── docs/             # API文档
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── README.md
│   │
│   ├── metadata-service/               # 元数据管理服务
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── handler/
│   │   │   │   ├── metadata.go        # 元数据CRUD API
│   │   │   │   ├── search.go          # 搜索API
│   │   │   │   ├── stats.go           # 统计API
│   │   │   │   └── health.go          # 健康检查
│   │   │   ├── service/
│   │   │   │   ├── metadata.go        # 元数据业务逻辑
│   │   │   │   ├── search.go          # 搜索功能
│   │   │   │   ├── stats.go           # 统计功能
│   │   │   │   └── cache.go           # 缓存管理
│   │   │   ├── repository/
│   │   │   │   ├── postgres.go        # PostgreSQL数据访问
│   │   │   │   └── redis.go           # Redis缓存访问
│   │   │   ├── middleware/
│   │   │   │   ├── error_injection.go # 错误注入
│   │   │   │   ├── tracing.go
│   │   │   │   ├── metrics.go
│   │   │   │   └── logging.go
│   │   │   ├── model/
│   │   │   │   ├── metadata.go        # 元数据模型
│   │   │   │   ├── request.go
│   │   │   │   └── response.go
│   │   │   └── config/
│   │   │       └── config.go
│   │   ├── migrations/                # 数据库迁移脚本
│   │   │   ├── 001_initial.up.sql
│   │   │   └── 001_initial.down.sql
│   │   ├── api/
│   │   │   └── swagger/
│   │   │       ├── metadata.yaml
│   │   │       └── docs/
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── README.md
│   │
│   ├── storage-service/                # 分布式存储服务(存储引擎层)
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── handler/
│   │   │   │   ├── storage.go         # 存储操作API
│   │   │   │   ├── replication.go     # 复制管理API
│   │   │   │   ├── node.go            # 节点管理API
│   │   │   │   ├── consistency.go     # 一致性检查API
│   │   │   │   └── health.go          # 健康检查
│   │   │   ├── service/
│   │   │   │   ├── storage.go         # 存储核心逻辑
│   │   │   │   ├── replication.go     # 多节点复制逻辑
│   │   │   │   ├── consistency.go     # 数据一致性管理
│   │   │   │   ├── recovery.go        # 数据恢复与修复
│   │   │   │   ├── third_party.go     # 第三方数据源集成
│   │   │   │   ├── node_manager.go    # 存储节点管理
│   │   │   │   └── load_balancer.go   # 存储负载均衡
│   │   │   ├── storage/
│   │   │   │   ├── filesystem.go      # 文件系统存储实现
│   │   │   │   ├── node.go            # 存储节点实现
│   │   │   │   ├── manager.go         # 存储管理器
│   │   │   │   ├── sharding.go        # 数据分片逻辑
│   │   │   │   └── compression.go     # 数据压缩
│   │   │   ├── middleware/
│   │   │   │   ├── error_injection.go # 存储错误注入
│   │   │   │   ├── failure_simulation.go # 磁盘/网络故障模拟
│   │   │   │   ├── latency_injection.go # 存储延迟注入
│   │   │   │   ├── capacity_limit.go  # 容量限制模拟
│   │   │   │   ├── tracing.go
│   │   │   │   ├── metrics.go
│   │   │   │   └── logging.go
│   │   │   ├── model/
│   │   │   │   ├── storage.go         # 存储操作模型
│   │   │   │   ├── node.go            # 节点状态模型
│   │   │   │   ├── replication.go     # 复制策略模型
│   │   │   │   ├── request.go
│   │   │   │   └── response.go
│   │   │   └── config/
│   │   │       └── config.go
│   │   ├── api/
│   │   │   └── swagger/
│   │   │       ├── storage.yaml
│   │   │       └── docs/
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── README.md
│   │
│   ├── async-service/                  # 异步任务处理服务
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── handler/
│   │   │   │   ├── queue.go           # 队列管理API
│   │   │   │   ├── task.go            # 任务管理API
│   │   │   │   ├── monitor.go         # 监控API
│   │   │   │   └── health.go          # 健康检查
│   │   │   ├── service/
│   │   │   │   ├── producer.go        # 消息生产者
│   │   │   │   ├── consumer.go        # 消息消费者
│   │   │   │   ├── processor.go       # 任务处理器
│   │   │   │   └── monitor.go         # 队列监控
│   │   │   ├── queue/
│   │   │   │   ├── redis_stream.go    # Redis Streams实现
│   │   │   │   ├── manager.go         # 队列管理器
│   │   │   │   └── worker.go          # Worker实现
│   │   │   ├── handlers/              # 具体任务处理器
│   │   │   │   ├── delete_handler.go  # 删除任务
│   │   │   │   ├── cleanup_handler.go # 清理任务
│   │   │   │   ├── sync_handler.go    # 同步任务
│   │   │   │   └── replication_handler.go # 复制任务
│   │   │   ├── client/
│   │   │   │   ├── storage.go         # 存储服务客户端
│   │   │   │   └── metadata.go        # 元数据服务客户端
│   │   │   ├── middleware/
│   │   │   │   ├── error_injection.go # 队列错误注入
│   │   │   │   ├── failure_simulation.go # 任务失败模拟
│   │   │   │   ├── tracing.go
│   │   │   │   ├── metrics.go
│   │   │   │   └── logging.go
│   │   │   ├── model/
│   │   │   │   ├── task.go            # 任务模型
│   │   │   │   ├── queue.go           # 队列模型
│   │   │   │   ├── request.go
│   │   │   │   └── response.go
│   │   │   └── config/
│   │   │       └── config.go
│   │   ├── api/
│   │   │   └── swagger/
│   │   │       ├── async.yaml
│   │   │       └── docs/
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── README.md
│   │
│   ├── error-injection/                # 错误注入控制服务
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── handler/
│   │   │   │   ├── injection.go       # 错误注入配置API
│   │   │   │   ├── scenario.go        # 错误场景管理API  
│   │   │   │   ├── monitoring.go      # 注入监控API
│   │   │   │   └── health.go
│   │   │   ├── service/
│   │   │   │   ├── injection.go       # 错误注入核心逻辑
│   │   │   │   ├── scenario.go        # 场景管理
│   │   │   │   ├── rules.go           # 注入规则引擎
│   │   │   │   └── monitoring.go      # 注入效果监控
│   │   │   ├── engine/
│   │   │   │   ├── http_errors.go     # HTTP错误注入
│   │   │   │   ├── network_errors.go  # 网络错误注入
│   │   │   │   ├── latency_injection.go # 延迟注入
│   │   │   │   ├── storage_errors.go  # 存储错误注入
│   │   │   │   └── chaos_monkey.go    # 混沌工程
│   │   │   ├── repository/
│   │   │   │   └── redis.go           # 注入配置存储
│   │   │   ├── middleware/
│   │   │   │   ├── tracing.go
│   │   │   │   ├── metrics.go
│   │   │   │   └── logging.go
│   │   │   ├── model/
│   │   │   │   ├── injection.go       # 注入规则模型
│   │   │   │   ├── scenario.go        # 场景模型
│   │   │   │   ├── request.go
│   │   │   │   └── response.go
│   │   │   └── config/
│   │   │       └── config.go
│   │   ├── scenarios/                 # 预定义错误场景
│   │   │   ├── network_partition.yaml
│   │   │   ├── storage_failure.yaml
│   │   │   ├── high_latency.yaml
│   │   │   └── service_overload.yaml
│   │   ├── api/
│   │   │   └── swagger/
│   │   │       ├── injection.yaml
│   │   │       └── docs/
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── README.md
│   │
│   ├── admin-api/                      # 管理API服务
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── handler/
│   │   │   │   ├── dashboard.go       # 仪表板API
│   │   │   │   ├── system.go          # 系统管理API
│   │   │   │   ├── monitoring.go      # 监控API
│   │   │   │   ├── logs.go            # 日志查询API
│   │   │   │   └── health.go
│   │   │   ├── service/
│   │   │   │   ├── dashboard.go       # 仪表板服务
│   │   │   │   ├── system.go          # 系统管理
│   │   │   │   ├── monitoring.go      # 监控数据聚合
│   │   │   │   └── logs.go            # 日志聚合
│   │   │   ├── client/                # 聚合其他服务数据
│   │   │   │   ├── s3_api.go
│   │   │   │   ├── metadata.go
│   │   │   │   ├── storage.go
│   │   │   │   ├── async.go
│   │   │   │   └── error_injection.go
│   │   │   ├── middleware/
│   │   │   │   ├── tracing.go
│   │   │   │   ├── metrics.go
│   │   │   │   └── logging.go
│   │   │   ├── model/
│   │   │   │   ├── dashboard.go       # 仪表板模型
│   │   │   │   ├── system.go          # 系统状态模型
│   │   │   │   ├── request.go
│   │   │   │   └── response.go
│   │   │   └── config/
│   │   │       └── config.go
│   │   ├── web/                       # Web静态资源（可选）
│   │   ├── api/
│   │   │   └── swagger/
│   │   │       ├── admin.yaml
│   │   │       └── docs/
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── README.md
│   │
│   └── config-service/                 # 配置管理服务
│       ├── cmd/
│       │   └── main.go
│       ├── internal/
│       │   ├── handler/
│       │   │   ├── config.go          # 配置管理API
│       │   │   ├── discovery.go       # 服务发现API
│       │   │   └── health.go
│       │   ├── service/
│       │   │   ├── config.go          # 配置管理
│       │   │   ├── discovery.go       # 服务发现
│       │   │   └── consul.go          # Consul集成
│       │   ├── repository/
│       │   │   └── consul.go          # Consul KV存储
│       │   ├── middleware/
│       │   │   ├── tracing.go
│       │   │   ├── metrics.go
│       │   │   └── logging.go
│       │   ├── model/
│       │   │   ├── config.go
│       │   │   ├── service.go         # 服务模型
│       │   │   ├── request.go
│       │   │   └── response.go
│       │   └── config/
│       │       └── config.go
│       ├── api/
│       │   └── swagger/
│       │       ├── config.yaml
│       │       └── docs/
│       ├── Dockerfile
│       ├── go.mod
│       └── README.md
│
├── shared/                              # 共享库和工具
│   │   ├── httpclient/                 # HTTP客户端工具包
│   │   │   ├── client.go              # HTTP客户端封装
│   │   │   ├── retry.go               # 重试机制
│   │   │   ├── circuit_breaker.go     # 熔断器
│   │   │   └── error_injection.go     # 客户端错误注入
│   │   ├── httpserver/                 # HTTP服务器工具包
│   │   │   ├── server.go              # 服务器封装
│   │   │   ├── middleware.go          # 通用中间件
│   │   │   └── shutdown.go
│   │   ├── database/
│   │   │   ├── postgres.go            # PostgreSQL连接池
│   │   │   ├── redis.go               # Redis连接池
│   │   │   └── transaction.go         # 事务管理
│   │   ├── discovery/
│   │   │   ├── consul.go              # Consul服务发现
│   │   │   └── health.go              # 健康检查
│   │   ├── config/
│   │   │   ├── loader.go              # 配置加载器
│   │   │   ├── consul.go              # Consul配置
│   │   │   └── env.go                 # 环境变量
│   │   ├── logger/
│   │   │   ├── logger.go              # 结构化日志
│   │   │   ├── elastic.go             # ElasticSearch日志
│   │   │   └── postgres.go            # PostgreSQL日志存储
│   │   ├── metrics/
│   │   │   ├── prometheus.go          # Prometheus指标
│   │   │   ├── counters.go            # 计数器指标
│   │   │   └── histograms.go          # 直方图指标
│   │   ├── tracing/
│   │   │   ├── tracer.go              # 链路追踪
│   │   │   └── elastic.go             # ElasticSearch追踪存储
│   │   ├── errors/
│   │   │   ├── types.go               # 错误类型定义
│   │   │   ├── injection.go           # 错误注入工具
│   │   │   └── handling.go            # 错误处理
│   │   ├── model/
│   │   │   ├── common.go              # 通用模型
│   │   │   ├── api.go                 # API通用模型
│   │   │   └── error.go               # 错误模型
│   │   └── utils/
│   │       ├── hash.go                # 哈希工具
│   │       ├── uuid.go                # UUID生成
│   │       ├── validator.go           # 输入验证
│   │       └── testing.go             # 测试工具
│   ├── go.mod
│   └── go.sum
│
├── infrastructure/                      # 基础设施配置
│   ├── consul/
│   │   ├── config/
│   │   │   └── consul.hcl             # Consul配置
│   │   ├── data/                      # 数据目录
│   │   └── Dockerfile
│   ├── postgres/
│   │   ├── init/
│   │   │   ├── 01-create-databases.sql
│   │   │   └── 02-init-users.sql
│   │   ├── config/
│   │   │   └── postgresql.conf
│   │   └── Dockerfile
│   ├── redis/
│   │   ├── config/
│   │   │   └── redis.conf
│   │   └── Dockerfile
│   ├── nginx/
│   │   ├── config/
│   │   │   ├── nginx.conf
│   │   │   └── conf.d/
│   │   └── Dockerfile
│   ├── prometheus/
│   │   ├── config/
│   │   │   ├── prometheus.yml
│   │   │   └── rules/
│   │   └── Dockerfile
│   └── elastic/
│       ├── config/
│       │   └── elasticsearch.yml
│       └── Dockerfile
│
├── deploy/                             # 部署配置
│
├── scripts/                            # 工具脚本
│   ├── build/
│   │   ├── build-all.sh              # 构建所有服务
│   │   └── docker-build.sh           # Docker构建脚本
│   ├── deploy/
│   │   ├── local-deploy.sh           # 本地部署
│   │   └── staging-deploy.sh         # 测试环境部署
│   ├── test/
│   │   ├── integration-test.sh       # 集成测试
│   │   └── chaos-test.sh             # 混沌测试
│   └── monitoring/
│       ├── health-check.sh           # 健康检查脚本
│       └── log-analysis.sh           # 日志分析脚本
│
├── tests/                              # 测试套件
│   ├── integration/                   # 集成测试
│   ├── chaos/                         # 混沌工程测试
│   └── e2e/                           # 端到端测试
│
├── docs/                               # 项目文档
│   ├── api/                           # API文档
│   │   ├── s3-api.md
│   │   ├── admin-api.md
│   │   └── error-injection.md
│   ├── architecture/                  # 架构文档
│   │   ├── overview.md
│   │   ├── microservices.md
│   │   └── error-simulation.md
│   ├── deployment/                    # 部署文档
│   │   ├── local-setup.md
│   │   ├── production-deployment.md
│   │   └── monitoring-setup.md
│   └── development/                   # 开发文档
│       ├── contributing.md
│       ├── coding-standards.md
│       └── testing-guide.md
│
└── tools/                             # 开发工具
    ├── chaos-monkey/                  # 混沌工程工具
    │   ├── cmd/
    │   └── scenarios/
    ├── load-generator/                # 负载生成器
    │   ├── cmd/
    │   └── profiles/
    └── log-parser/                    # 日志解析工具
        ├── cmd/
        └── parsers/
```