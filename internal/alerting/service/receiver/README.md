🧭 端到端验证（Docker Postgres + Redis + 本服务）

以下步骤演示从 Alertmanager Webhook 到数据库落库的完整链路验证：

1) 启动 Postgres（Docker）

```bash
docker run --name zeroops-pg \
  -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=zeroops \
  -p 5432:5432 -d postgres:16
```

1b) 启动 Redis（Docker）

```bash
docker run --name zeroops-redis -p 6379:6379 -d redis:7-alpine
```

2) 初始化告警相关表
运行集成测试（需 Postgres 实例与 `-tags=integration`）可验证插入成功：
```bash
go test ./internal/alerting/service/receiver -tags=integration -run TestPgDAO_InsertAlertIssue -v
```

3) 配置环境变量并启动服务（另开一个 shell 后台运行）

```bash
export DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASSWORD=postgres DB_NAME=zeroops DB_SSLMODE=disable
export ALERT_WEBHOOK_BASIC_USER=alert ALERT_WEBHOOK_BASIC_PASS=REDACTED
export REDIS_ADDR=localhost:6379 REDIS_PASSWORD="" REDIS_DB=0
nohup go run ./cmd/zeroops -- 1>/tmp/zeroops.out 2>&1 &
```

4) 用 curl 模拟 Alertmanager 发送 firing 事件

```bash
curl -u alert:REDACTED -H 'Content-Type: application/json' \
  -X POST http://localhost:8080/v1/integrations/alertmanager/webhook -d '{
  "receiver":"our-webhook",
  "status":"firing",
  "alerts":[{
    "status":"firing",
    "labels":{"alertname":"HighRequestLatency","service":"serviceA","severity":"P1","idc":"yzh"},
    "annotations":{"summary":"p95 latency over threshold","description":"apitime p95 > 450ms"},
    "startsAt":"2025-05-05T11:00:00Z",
    "endsAt":"0001-01-01T00:00:00Z",
    "generatorURL":"http://prometheus/graph?g0.expr=...",
    "fingerprint":"3b1b7f4e8f0e"
  }],
  "groupLabels":{"alertname":"HighRequestLatency"},
  "commonLabels":{"service":"serviceA","severity":"P1"},
  "version":"4"
}'
```

5) 在数据库中验证（应看到一行 Open/P1/Pending 且标题匹配的记录）

```bash
docker exec -i zeroops-pg psql -U postgres -d zeroops -c \
  "SELECT id,state,level,alert_state,title,alert_since FROM alert_issues WHERE title='p95 latency over threshold' AND alert_since='2025-05-05 11:00:00'::timestamp;"
```
```bash
# 更易读（格式化 JSON）labels
docker exec -i zeroops-pg psql -U postgres -d zeroops -c \
   "SELECT jsonb_pretty(labels::jsonb) AS label FROM alert_issues WHERE title='p95 latency over threshold' AND alert_since='2025-05-05 11:00:00'::timestamp;"

```

6)（可选）运行带集成标签的最小 DAO 测试

```bash
DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASSWORD=postgres DB_NAME=zeroops DB_SSLMODE=disable \
go test ./internal/alerting/service/receiver -tags=integration -run TestPgDAO_InsertAlertIssue -v
```


receiver/ — 从 Alertmanager Webhook 到 alert_issues 入库的实施计划

目标：当 Alertmanager 向本服务发起 POST JSON 时，第一次创建告警记录并落表 alert_issues，字段规则：
	•	state 默认 Open
	•	alertState 默认 Pending
	•	其余字段按 webhook 请求体解析、校验后写入

本计划仅覆盖「首次创建」逻辑；resolved（恢复）更新逻辑可在后续补充（例如切换 state=Closed、alertState=Restored）。

⸻

① 目录与文件准备

在 alerting/service/receiver/ 下新建如下文件（按模块职责划分）：

alerting/
└─ service/
   └─ receiver/
      ├─ README.md                 # ← 就放本文档
      ├─ router.go                 # 注册路由：POST /v1/integrations/alertmanager/webhook
      ├─ handler.go                # HTTP 入口，接收与整体编排
      ├─ dto.go                    # 入参（Alertmanager Webhook）与内部 DTO 定义
      ├─ validator.go              # 字段校验（必填/枚举/时间格式等）
      ├─ mapper.go                 # 映射：AM payload → alert_issues 行记录
      ├─ dao.go                    # DB 访问（Insert/Query/事务/重试）
      ├─ cache.go                  # Redis 客户端与写通缓存（Write-through）
      ├─ idempotency.go            # 幂等键生成与“已处理”快速判断（应用层）
      └─ errors.go                 # 统一错误定义（参数错误/DB错误等）

若你的 DB 连接封装在 alerting/database/，dao.go 里直接引入公用的 db 客户端即可。

⸻

② 路由与入口

router.go

// package receiver
func RegisterReceiverRoutes(r *gin.Engine, h *Handler) {
    r.POST("/v1/integrations/alertmanager/webhook", h.AlertmanagerWebhook)
}

handler.go

type Handler struct {
    dao *DAO
    cache *Cache // Redis 写通
}

func NewHandler(dao *DAO, cache *Cache) *Handler { return &Handler{dao: dao, cache: cache} }

func (h *Handler) AlertmanagerWebhook(c *gin.Context) {
    var req AMWebhook // dto.go 中定义的 Alertmanager 请求体结构
    if err := c.BindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "invalid JSON"})
        return
    }

    // 1) 基本字段校验
    if err := ValidateAMWebhook(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
        return
    }

    // 2) 仅处理 status == "firing" 的首次创建
    if strings.ToLower(req.Status) != "firing" {
        c.JSON(http.StatusOK, gin.H{"ok": true, "msg": "ignored (not firing)"})
        return
    }

    // 3) 对每条 alert 做落库（可能一批多个）
    //    幂等键建议：fingerprint + startsAt（同一告警起始时间视为同一事件）
    created := 0
    for _, a := range req.Alerts {
        key := BuildIdempotencyKey(a)         // idempotency.go
        if AlreadySeen(key) {                 // 应用层短路（可选）
            continue
        }

        row, mapErr := MapToAlertIssueRow(&req, &a) // mapper.go → 组装 alert_issues 行
        if mapErr != nil {
            // 单条失败不影响其它，记录日志即可
            continue
        }

        // 4) 插入 DB（第一次创建强制 state=Open, alertState=Pending）
        if err := h.dao.InsertAlertIssue(c, row); err != nil {
            // 若唯一约束冲突/网络抖动等，记录后继续
            continue
        }
        // 5) 同步写入 service_states（health_state=Error；detail/resolved_at/correlation_id 留空）
        //    service 从 labels.service 取；version 可从 labels.service_version 取（可空）
        if err := h.dao.UpsertServiceState(c, a.Labels["service"], a.Labels["service_version"], row.AlertSince, "Error"); err != nil {
            // 仅记录错误，不阻断主流程
        }
        // 6) 写通到 Redis（不阻塞主流程，失败仅记录日志）
        //    alert_issues
        if err := h.cache.WriteIssue(c, row, a); err != nil {
            // 仅记录错误，避免影响 Alertmanager 重试逻辑
        }
        //    service_states
        _ = h.cache.WriteServiceState(c, a.Labels["service"], a.Labels["service_version"], row.AlertSince, "Error")
        MarkSeen(key) // 记忆幂等键
        created++
    }

    c.JSON(http.StatusOK, gin.H{"ok": true, "created": created})
}


⸻

③ 入参 DTO 与内部结构

dto.go（Alertmanager Webhook 载荷 + 内部插入结构）

type KV map[string]string

// 来自 Alertmanager 的单条告警
type AMAlert struct {
    Status       string    `json:"status"`       // firing|resolved
    Labels       KV        `json:"labels"`       // 包含 alertname、service、severity 等
    Annotations  KV        `json:"annotations"`  // 包含 summary/description 等
    StartsAt     time.Time `json:"startsAt"`
    EndsAt       time.Time `json:"endsAt"`
    GeneratorURL string    `json:"generatorURL"`
    Fingerprint  string    `json:"fingerprint"`  // 用于幂等
}

// Webhook 根对象
type AMWebhook struct {
    Receiver          string    `json:"receiver"`
    Status            string    `json:"status"`            // firing|resolved
    Alerts            []AMAlert `json:"alerts"`
    GroupLabels       KV        `json:"groupLabels"`
    CommonLabels      KV        `json:"commonLabels"`
    CommonAnnotations KV        `json:"commonAnnotations"`
    ExternalURL       string    `json:"externalURL"`
    Version           string    `json:"version"`
    GroupKey          string    `json:"groupKey"`
}

// 准备插入 alert_issues 的行（与表字段一一对应）
type AlertIssueRow struct {
    ID         string          // uuid
    State      string          // enum: Open/Closed （首次固定 Open）
    Level      string          // varchar(32): P0/P1/P2/Warning
    AlertState string          // enum: Pending/InProcessing/Restored/AutoRestored（首次固定 Pending）
    Title      string          // varchar(255)
    LabelJSON  json.RawMessage // json: 标准化后的 [{key,value}]
    AlertSince time.Time       // timestamp: 用 StartsAt
}


⸻

④ 字段校验（validator）

validator.go

func ValidateAMWebhook(w *AMWebhook) error {
    if w == nil { return errors.New("nil payload") }
    if len(w.Alerts) == 0 { return errors.New("alerts empty") }
    // 可加大小限制：len(alerts) <= N；防巨量 payload
    for i := range w.Alerts {
        a := &w.Alerts[i]
        if a.StartsAt.IsZero() { return fmt.Errorf("alerts[%d].startsAt empty", i) }
        // 允许空 annotations.summary，但后续会用回退规则生成 title
        if a.Status == "" { a.Status = "firing" } // 容错
    }
    return nil
}

var allowedLevels = map[string]bool{"P0":true,"P1":true,"P2":true,"Warning":true}

func NormalizeLevel(sev string) string {
    s := strings.ToUpper(strings.TrimSpace(sev))
    if allowedLevels[s] { return s }
    // 若为空/不合法，可设置默认或交给 severity 模块再评估
    return "Warning"
}


⸻

⑤ 映射规则（mapper）

目标：将 Alertmanager 的单条 AMAlert → AlertIssueRow。
	•	id：uuid.NewString()
	•	state：Open（首次创建强制）
	•	alertState：InProcessing（首次创建强制）
	•	level：NormalizeLevel(alert.Labels["severity"])
	•	title：优先 annotations.summary，否则拼：{idc} {service} {alertname} ...
	•	label：把 labels 展平成 [{key,value}]（额外加上一些关键来源信息：am_fingerprint、generatorURL、groupKey）
	•	alertSince：StartsAt（统一转 UTC）

mapper.go

func MapToAlertIssueRow(w *AMWebhook, a *AMAlert) (*AlertIssueRow, error) {
    // 1) Title
    title := strings.TrimSpace(a.Annotations["summary"])
    if title == "" {
        // fallback：尽量信息量大且≤255
        title = fmt.Sprintf("%s %s %s",
            a.Labels["idc"], a.Labels["service"], a.Labels["alertname"])
        title = strings.TrimSpace(title)
        if title == "" { title = "Alert from Alertmanager" }
    }
    if len(title) > 255 { title = title[:255] }

    // 2) Level
    level := NormalizeLevel(a.Labels["severity"])

    // 3) Labels → []{key,value}
    //    附加指纹等方便后续查询/对账
    flat := make([]map[string]string, 0, len(a.Labels)+3)
    for k, v := range a.Labels {
        flat = append(flat, map[string]string{"key": k, "value": v})
    }
    if a.Fingerprint != "" {
        flat = append(flat, map[string]string{"key": "am_fingerprint", "value": a.Fingerprint})
    }
    if g := strings.TrimSpace(a.GeneratorURL); g != "" {
        flat = append(flat, map[string]string{"key": "generatorURL", "value": g})
    }
    if w.GroupKey != "" {
        flat = append(flat, map[string]string{"key": "groupKey", "value": w.GroupKey})
    }
    b, _ := json.Marshal(flat)

    // 4) Row
    return &AlertIssueRow{
        ID:         uuid.NewString(),
        State:      "Open",
        AlertState: "Pending",
        Level:      level,
        Title:      title,
        LabelJSON:  b,
        AlertSince: a.StartsAt.UTC(), // 建议统一 UTC
    }, nil
}


⸻

⑥ 幂等（idempotency）

虽然本步骤主要描述“首次创建”，但为了避免重复插入，建议引入应用层幂等（无须改表结构）：

idempotency.go

func BuildIdempotencyKey(a AMAlert) string {
    return a.Fingerprint + "|" + a.StartsAt.UTC().Format(time.RFC3339Nano)
}

// 可以用内存 LRU/Redis；或入库前先按 (am_fingerprint + startsAt) 查询是否存在
func AlreadySeen(key string) bool { /* TODO */ return false }
func MarkSeen(key string)         { /* TODO */ }

若后续允许调整表结构，可把 am_fingerprint 单列化并与 alertSince 组成唯一索引，幂等更稳。

⸻

⑦ 数据访问（DAO）

dao.go（示例使用 pgx / database/sql，重点是参数化与事务）

type DAO struct{ DB *pgxpool.Pool }

func (d *DAO) InsertAlertIssue(ctx context.Context, r *AlertIssueRow) error {
    const q = `
    INSERT INTO alert_issues
        (id, state, level, alert_state, title, labels, alert_since)
    VALUES
        ($1, $2, $3, $4, $5, $6, $7)
    `
    _, err := d.DB.Exec(ctx, q,
        r.ID, r.State, r.Level, r.AlertState, r.Title, r.LabelJSON, r.AlertSince)
    return err
}

注意：
	•	label 列类型为 json（建议实际使用 jsonb），此处用 json.RawMessage 参数化写入即可。
	•	使用 Exec/Prepare 都可，确保不拼接字符串，防注入。
	•	生产建议增加：重试策略、插入耗时监控、错误分级（唯一冲突 vs 网络抖动）。

⸻

⑧ Redis 缓存写通（Write-through）与分布式幂等

目标：在成功写入 PostgreSQL 后，将关键数据写入 Redis，既为前端查询提供加速缓存，也为后续定时任务提供快速读取能力；同时用 Redis 提供跨实例幂等控制。

依赖：

```bash
go get github.com/redis/go-redis/v9
```

配置（环境变量）：

```
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=""
REDIS_DB=0
```

key 设计与 TTL：

- alert:issue:{id} → JSON（AlertIssueRow + 补充字段），TTL 3d
- alert:idemp:{fingerprint}|{startsAtRFC3339Nano} → "1"，TTL 10m（用于分布式幂等 SETNX）
- alert:index:open → Set(issues...)，无 TTL（恢复时再移除）
- alert:index:svc:{service}:open → Set(issues...)，无 TTL
// service_states 缓存
- service_state:{service}:{version} → JSON（service/version/report_at/health_state），TTL 3d
- service_state:index:service:{service} → Set(keys)
- service_state:index:health:{health_state} → Set(keys)

cache.go（示例）：

```go
type Cache struct{ R *redis.Client }

func NewCacheFromEnv() *Cache {
    db, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
    c := redis.NewClient(&redis.Options{Addr: os.Getenv("REDIS_ADDR"), Password: os.Getenv("REDIS_PASSWORD"), DB: db})
    return &Cache{R: c}
}

// 写通：issue 主键对象 + 索引集合
func (c *Cache) WriteIssue(ctx context.Context, r *AlertIssueRow, a AMAlert) error {
    if c == nil || c.R == nil { return nil }
    key := "alert:issue:" + r.ID
    payload := map[string]any{
        "id": r.ID, "state": r.State, "level": r.Level, "alertState": r.AlertState,
        "title": r.Title, "labels": json.RawMessage(r.LabelJSON), "alertSince": r.AlertSince,
        "fingerprint": a.Fingerprint, "service": a.Labels["service"], "alertname": a.Labels["alertname"],
    }
    b, _ := json.Marshal(payload)
    svc := strings.TrimSpace(a.Labels["service"])
    pipe := c.R.Pipeline()
    pipe.Set(ctx, key, b, 72*time.Hour)
    pipe.SAdd(ctx, "alert:index:open", r.ID)
    if svc != "" {
        pipe.SAdd(ctx, "alert:index:svc:"+svc+":open", r.ID)
    }
    _, err := pipe.Exec(ctx)
    return err
}

// 分布式幂等：SETNX + TTL
func (c *Cache) TryMarkIdempotent(ctx context.Context, a AMAlert) (bool, error) {
    if c == nil || c.R == nil { return true, nil }
    k := "alert:idemp:" + a.Fingerprint + "|" + a.StartsAt.UTC().Format(time.RFC3339Nano)
    ok, err := c.R.SetNX(ctx, k, "1", 10*time.Minute).Result()
    return ok, err
}
```

在 handler 中接入（伪码）：

```go
// 幂等短路（跨实例）
if ok, _ := h.cache.TryMarkIdempotent(c, a); !ok {
    continue
}
// DB 成功后写通 Redis
_ = h.cache.WriteIssue(c, row, a)
```

失败处理：Redis 失败不影响 HTTP 主流程（Alertmanager 侧重试依赖 2xx），但需要日志打点与告警；后续可在定时任务做补偿（扫描最近 N 分钟的 DB 记录回填 Redis）。

快速验证：

```bash
# 触发一次 webhook 后在 Redis 查看
redis-cli --raw keys 'alert:*'
redis-cli --raw get alert:issue:<id>
redis-cli --raw smembers alert:index:open | head -n 10
redis-cli ttl alert:issue:<id>
redis-cli --raw keys 'service_state:*'
redis-cli --raw get service_state:serviceA:v1.3.7
redis-cli --raw smembers service_state:index:health:Error
```

⸻

⑨ 成功/失败返回与日志
	•	返回：统一 200 {"ok": true, "created": <n>}，即使个别记录失败也快速返回，避免 Alertmanager 阻塞重试。
	•	日志：按 alertname/service/severity/fingerprint 打点；错误包含 SQLSTATE/堆栈；统计接收/解析/插入耗时分位。

⸻

⑩ 最小联调（人工模拟）

firing 模拟：

curl -X POST http://localhost:8080/v1/integrations/alertmanager/webhook \
  -H 'Content-Type: application/json' \
  -d '{
    "receiver":"our-webhook",
    "status":"firing",
    "alerts":[
      {
        "status":"firing",
        "labels":{
            "alertname":"HighRequestLatency",
            "service":"serviceA",
            "severity":"P1",
            "idc":"yzh",
            "service_version": "v1.3.7"
            },
        "annotations":{"summary":"p95 latency over threshold","description":"apitime p95 > 450ms"},
        "startsAt":"2025-05-05T11:00:00Z",
        "endsAt":"0001-01-01T00:00:00Z",
        "generatorURL":"http://prometheus/graph?g0.expr=...",
        "fingerprint":"3b1b7f4e8f0e"
      }
    ],
    "groupLabels":{"alertname":"HighRequestLatency"},
    "commonLabels":{"service":"serviceA","severity":"P1"},
    "version":"4"
  }'

入库后，alert_issues 里应看到：
	•	state=Open
	•	alertState=Pending
	•	level=P1
	•	title="p95 latency over threshold"
	•	label 中包含 am_fingerprint/generatorURL/groupKey/...
	•	alertSince=2025-05-05 11:00:00+00

同时，service_states 里应看到/更新（按 service+version）：
	•	service=serviceA
	•	version=（若 labels 中有 service_version 则为其值，否则为空字符串）
	•	report_at=与 alert_since 一致（若已存在则保留更早的 report_at）
	•	health_state=Error
	•	detail/resolved_at/correlation_id 为空

Redis 中应看到：
	•	key: alert:issue:<id> 值为 JSON 且 TTL≈3 天
	•	集合 alert:index:open 中包含 <id>
	•	若有 service=serviceA，则 alert:index:svc:serviceA:open 包含 <id>

⸻