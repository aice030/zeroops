## Mock S3 Storage 微服务开发路线

本路线用于将本项目改造为微服务架构，重点在于“最大化可注入的错误场景”，以支撑可靠性与故障演练。本文覆盖从本地环境、共享底座、契约驱动、最小闭环，到异步、错误注入、网关、可观测、配置中心、测试与交付的全流程。项目不需要鉴权。

### 目标与范围

- 将单体能力拆分为可独立扩展的服务：`s3-api`（编排层）、`storage-service`（存储引擎层）、`metadata-service`（元数据）、`async-service`（异步/队列）、`error-injection`（错误注入控制）、`gateway`（Nginx）、`admin-api`（聚合管理）、`config-service`（配置中心）。
- 对外兼容 S3 常用接口（PUT/GET/DELETE/HEAD/List），并提供管理与注入 API。
- 在服务端和客户端关键路径植入可控错误注入点，支持多种故障组合。
- 无鉴权，便于测试与演示。

### 技术栈（参考根仓库 README）

- 容器化：Docker / docker-compose（可拓展 Helm）
- 服务通信：HTTP（保留 gRPC 备选）
- 服务发现：Consul
- API 网关：Nginx
- 数据层：PostgreSQL + Redis（缓存/队列），后续可换 Kafka
- 配置中心：Consul KV
- 观测（监控/日志/链路）：OpenTelemetry（SDK/Collector）；日志与链路数据存储到 Elasticsearch（可配 Kibana 可视化）

---

## 分阶段开发路线

每个阶段包含目标、主要动作与交付物。建议严格按顺序推进，优先打通最小闭环，再增强与扩展注入能力。

### 1) 基础设施与本地环境

- 目标：一键拉起依赖。
- 动作：在仓库根配置 `docker-compose.yml` 与 `.env`，组件包含：`consul`、`postgres`、`redis`、`opentelemetry-collector`、`elasticsearch`、`nginx`。
- 交付：`infrastructure/*` 配置就绪；`scripts/deploy/local-deploy.sh` 一键启停。

### 2) 共享库 `shared/pkg` 底座

- 目标：统一 Server/Client 能力，减少样板代码，并预埋错误注入钩子。
- 动作：
  - `httpserver`：请求ID、结构化日志、恢复、CORS、Prom 指标、Tracing 中间件。
  - `httpclient`：超时、重试、熔断、Tracing、指标，以及“错误注入钩子”。
  - `config`：环境变量+Consul KV 加载（支持定时刷新占位）。
  - `discovery`：Consul 注册与发现。
  - `metrics`、`tracing`、`logger`、`errors`（标准错误类型）。
- 交付：各服务可零样板接入统一 Server/Client 与观测。

### 3) 契约优先（OpenAPI/Swagger）

- 目标：明确接口，确保服务间边界清晰。
- 动作：在 `services/*/api/swagger/` 下定义：`s3-api.yaml`、`metadata.yaml`、`storage.yaml`、`async.yaml`、`injection.yaml`、`admin.yaml`、`config.yaml`；标注错误码与“可注入错误位点”。
- 交付：契约评审通过并冻结 v1。

### 4) 脚手架与健康检查

- 目标：服务可启动、可注册、可观测。
- 动作：各服务建立 `cmd/main.go` 与 `internal/{handler,service,repository,middleware,config}`；接入共享库；暴露 `/health`、OpenTelemetry 指标/日志/链路（OTLP/HTTP 或 gRPC，经 Collector 转发）；注册到 Consul。
- 交付：服务启动无报错，在 Consul 可见，OpenTelemetry Collector 可接收指标/日志/链路数据。

### 5) 最小闭环（写路径优先）

- 目标：打通 S3 PUT → 存储写入 → 元数据落库 → 投递异步任务。
- 动作：
  - `storage-service`（最小版）：文件系统实现（写/读/删 API），节点ID与路径由配置中心下发。
  - `metadata-service`（最小版）：PostgreSQL + Redis 缓存元数据。
  - `s3-api`（PUT）：校验→调用存储写入→写元数据→投递 `upload_completed` 至 `async-service`（先打桩）。
- 交付：`PUT /:bucket/:key` 端到端成功，打印/指标可见。

### 6) 读路径与列举

- 目标：完善 GET/HEAD/List；预留第三方回源接口。
- 动作：`s3-api` 通过元数据定位并调用存储读；List 提供分页；`storage-service` 设计回源占位（策略、节流、熔断、写回）。
- 交付：`GET/HEAD /:bucket/:key`、`GET /:bucket` 可用。

### 7) 异步服务初版 `async-service`

- 目标：任务入队/消费闭环。
- 动作：使用 Redis Streams 实现生产者、消费者、worker；实现 `upload_completed`、`delete_from_storage` 处理器；暴露消费指标与 lag 指标。
- 交付：PUT 后能看到 `upload_completed` 被消费；删除能下发 delete 任务。

### 8) 删除路径与异步删除

- 目标：DELETE 合规与异步清理。
- 动作：`s3-api` 先删元数据，再投递删除任务；`storage-service` 实做跨节点删除（≥1 节点成功即视为成功）。
- 交付：`DELETE /:bucket/:key` 可用，节点文件被清理。

### 9) 错误注入能力 `error-injection`（核心）

- 目标：可控、可组合的故障模拟能力。
- 动作：
  - 规则模型与 API：注入类型（HTTP 状态覆盖、延迟注入、错误率、网络超时、部分失败、容量限制、校验失败、缓存不一致、第三方回源失败等），支持场景/租期/命中率。
  - 共享库钩子：`httpserver`/`httpclient` 在处理前查询注入规则（本地缓存 Redis，规则源来自注入服务）。
  - 植入位点：
    - `s3-api`（编排超时/重试/回退/部分成功）；
    - `storage-service`（读/写/删 I/O、节点不可用、半失败、一致性偏差、回源失败/脏数据）；
    - `metadata-service`（DB 慢、事务失败、缓存穿透/击穿/雪崩）；
    - `async-service`（消息丢失、重复消费、死信）。
- 交付：通过注入 API 远程切换场景，E2E 可见指标/日志/链路的变化。

### 10) 网关 `gateway`（Nginx）

- 目标：统一入口与路由。
- 动作：`s3-api.conf`、`admin-api.conf` 路由；上游静态或经 Consul；可选基础限流（仍无鉴权）。
- 交付：外部仅经网关访问；路由正确、观测齐全。

### 11) 可观测与监控

- 目标：“看得见”每一种故障与恢复。
- 动作：
  - OpenTelemetry SDK 全面埋点，统一输出到 OpenTelemetry Collector。
  - 指标：QPS、P95/P99、错误率、重试次数、熔断状态、入/出队速率与积压、节点可用率、注入命中率等（经 OTEL 导出，可对接 Grafana/Prom 等）。
  - 链路：跨服务 Trace（包含注入规则/场景标签），由 Collector 写入 Elasticsearch。
  - 日志：结构化日志（包含 `scenario_id`、`inject_point`、`rule_id` 等字段），由 Collector 写入 Elasticsearch。
- 交付：可在 Kibana 中查看日志与链路，在 Grafana 中查看指标（或使用其他与 OTEL 兼容的指标后端）。

### 12) 管理聚合 `admin-api`

- 目标：一站式观测与运维入口。
- 动作：聚合各服务健康、统计、注入状态、队列监控、存储节点状态；输出 Dashboard 数据模型。
- 交付：管理接口与可视化就绪（简版可仅 JSON）。

### 13) 配置中心 `config-service`

- 目标：集中与动态配置。
- 动作：Consul KV 读写 API；共享库 `config` 支持启动加载与定时刷新；服务注册/健康检查统一化。
- 交付：无需重发镜像即可动态调参（限流、重试、注入策略等）。

### 14) 测试矩阵（集成/混沌/E2E/压测）

- 目标：可重复验证。
- 动作：
  - `tests/integration`：端到端功能用例；
  - `tests/chaos`：常见注入场景脚本；
  - `tools/load-generator`：上传/下载/列举/删除压测；
  - 覆盖：强/最终一致性、部分节点失败、网络隔离、队列积压、回源失败等。
- 交付：一键跑测试脚本，输出报告与指标快照。

### 15) 第三方回源完善（可选增强）

- 目标：真实回源策略与治理。
- 动作：在 `storage-service` 落地回源：缓存、节流、熔断、幂等写回；同时支持注入“第三方不可用/慢/脏数据”。
- 交付：GET 本地缺失时稳定回源并可被注入破坏。

### 16) 打包与交付（CI/CD）

- 目标：可持续发布。
- 动作：服务级 `Dockerfile`、`Makefile`；`scripts/build/build-all.sh`；Compose/Helm；版本与变更日志；基础回滚策略。
- 交付：本地/测试环境一键发版。

---

## 服务与目录对应（参考 README）

- `services/s3-api`：对外 S3 API（编排层），中间件最全，注入位点最丰富。
- `services/storage-service`：存储引擎层；节点管理、复制、一致性、回源、容量/延迟/失败注入。
- `services/metadata-service`：元数据 CRUD/搜索/统计；PostgreSQL + Redis 缓存。
- `services/async-service`：异步任务与 Worker；Redis Streams（可替 Kafka）。
- `services/error-injection`：错误注入规则管理与场景控制。
- `services/admin-api`：管理聚合与仪表盘数据。
- `services/config-service`：配置中心与服务发现。
- `services/gateway`：Nginx 网关与路由。

---

## 错误注入设计要点

- 注入范围：HTTP 级别（状态/延迟/超时）、存储 I/O（读写删失败/半成功/数据损坏/容量限制）、DB（慢查询/事务失败）、缓存（击穿/穿透/不一致）、消息队列（丢失/重复/死信）、第三方回源（不可用/慢/脏数据）。
- 注入策略：按服务/端点/百分比/租期/灰度/场景标签；支持组合规则与优先级。
- 执行位置：共享库 `httpserver`/`httpclient` 钩子 + 业务关键点（例如写入落盘前后、回源前后、提交事务前后）。
- 配置下发：`error-injection` API -> Redis 缓存 -> 各实例本地缓存；变更事件触发刷新。
- 安全措施：默认关闭；仅测试/预发环境启用；提供一键禁用与 TTL。

---

## 可观测性与指标建议

- 服务通用：`http_requests_total`、`http_request_duration_seconds`、`http_inflight_requests`、`errors_total`。
- 重试/熔断：`retry_attempts_total`、`circuit_breaker_open_total`。
- 队列：`enqueue_total`、`dequeue_total`、`consumer_lag`、`dead_letters_total`。
- 存储：节点可用率、I/O 时延、写放大、空间使用、回源命中率与失败率。
- 注入：规则命中率、按场景分布、覆盖率。
- Tracing：跨服务 span，携带 `scenario_id/rule_id` 标签。

---

## 本地启动与快速验证（示例）

1) 启动依赖与服务
```
scripts/deploy/local-deploy.sh   # 或 docker-compose up -d
```
2) 快速验证 S3 最小闭环
```
curl -X PUT  "http://localhost:8080/my-bucket/hello.txt" -H "Content-Type: text/plain" -d "Hello, World!"
curl -X GET  "http://localhost:8080/my-bucket/hello.txt"
curl -X HEAD "http://localhost:8080/my-bucket/hello.txt" -i
curl -X GET  "http://localhost:8080/my-bucket"
curl -X DELETE "http://localhost:8080/my-bucket/hello.txt"
```
3) 注入延迟示例（伪 API）
```
curl -X POST "http://localhost:9000/api/v1/injection/rules" -H "Content-Type: application/json" -d '{
  "service": "s3-api",
  "endpoint": "PUT /:bucket/:key",
  "type": "latency",
  "value_ms": 500,
  "rate": 0.5,
  "ttl": "5m",
  "scenario": "half_slow_write"
}'
```
